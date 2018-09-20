////////////////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 The mjoy-go Authors.
//
// The mjoy-go is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
// @File: backend.go
// @Date: 2018/05/08 18:02:08
////////////////////////////////////////////////////////////////////////////////

// Package mjoy implements the Mjoy protocol.
package mjoy

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"mjoy.io/consensus"

	"mjoy.io/node/services/mjoy/downloader"
	"mjoy.io/utils/event"
	"mjoy.io/node"
	"mjoy.io/params"
	"mjoy.io/communication/p2p"
	"mjoy.io/communication/rpc"
	"mjoy.io/utils/bloom"
	"mjoy.io/accounts"
	"mjoy.io/utils/database"
	"mjoy.io/core/blockchain"
	"mjoy.io/core/blockchain/block"
	"mjoy.io/common/types"
	"mjoy.io/core/txprocessor"
	"mjoy.io/core/chainindexer"
	//"mjoy.io/core/genesis"


	"mjoy.io/core/genesis"
	"mjoy.io/blockproducer"
	"mjoy.io/communication/rpc/mjoyapi"
	"mjoy.io/mjoyd/config"
	"mjoy.io/accounts/keystore"
	"crypto/ecdsa"
	"mjoy.io/core/interpreter"
)

type LesServer interface {
	Start(srvr *p2p.Server)
	Stop()
	Protocols() []p2p.Protocol
	SetBloomBitsIndexer(bbIndexer *chainindexer.ChainIndexer)
}

// Mjoy implements the Mjoy full node service.
type Mjoy struct {
	config      *Config
	chainConfig *params.ChainConfig

	// Channel for shutting down the service
	shutdownChan  chan bool    // Channel for shutting down the mjoy
	stopDbUpgrade func() error // stop chain db sequential key upgrade

	// Handlers
	txPool          *txprocessor.TxPool
	blockchain      *blockchain.BlockChain
	protocolManager *ProtocolManager
	lesServer       LesServer

	// DB interfaces
	chainDb database.IDatabase // Block chain database

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	bloomRequests chan chan *bloom.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *chainindexer.ChainIndexer             // Bloom indexer operating during block imports

	ApiBackend *MjoyApiBackend


	blockproducer     *blockproducer.Blockproducer
	interVm             *interpreter.Vms
	coinbase types.Address

	networkId     uint64
	//netRPCService *mjoyapi.PublicNetAPI

	lock sync.RWMutex // Protects the variadic fields (e.g. coinbase)
}

func (s *Mjoy) AddLesServer(ls LesServer) {
	s.lesServer = ls
	ls.SetBloomBitsIndexer(s.bloomIndexer)
}
type SetupGenesisResult struct {
	ChainConfig *params.ChainConfig
	GennesisHash *types.Hash
	GenesisErr error
	ChainDb  *database.IDatabase

}

// New creates a new Mjoy object (including the
// initialisation of the common Mjoy object)----Move to mjoy2.go
/**/
func New(ctx *node.ServiceContext) (*Mjoy, error) {
	c := config.GetConfigInstance()
	var config = &Config{}
	err := c.Register("mjoy", config)
	if err != nil {
		logger.Error("get config fail", "err", err)
	}

	if config.SyncMode == downloader.LightSync {
		return nil, errors.New("can't run mjoy.Mjoy in light sync mode, use les.LightMjoy")
	}
	if !config.SyncMode.IsValid() {
		return nil, fmt.Errorf("invalid sync mode %d", config.SyncMode)
	}
	chainDb, err := CreateDB(ctx, config, "chaindata")
	if err != nil {
		return nil, err
	}
	stopDbUpgrade := upgradeDeduplicateData(chainDb)
	chainConfig, genesisHash, genesisErr := genesis.SetupGenesisBlock(chainDb, config.Genesis)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}
	logger.Info("Initialised chain configuration", "config", chainConfig)

	mjoy := &Mjoy{
		config:         config,
		chainDb:        chainDb,
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		accountManager: ctx.AccountManager,
		shutdownChan:   make(chan bool),
		stopDbUpgrade:  stopDbUpgrade,
		networkId:      config.NetworkId,
		coinbase:      config.Coinbase,
		bloomRequests:  make(chan chan *bloom.Retrieval),
		bloomIndexer:   chainindexer.NewBloomIndexer(chainDb, params.BloomBitsBlocks),
	}

	logger.Info("Initialising Mjoy protocol", "versions", ProtocolVersions, "network", config.NetworkId)
	mjoy.engine = CreateConsensusEngine(mjoy)
	if !config.SkipBcVersionCheck {
		bcVersion := blockchain.GetBlockChainVersion(chainDb)
		if bcVersion != blockchain.BlockChainVersion && bcVersion != 0 {
			return nil, fmt.Errorf("Blockchain DB version mismatch (%d / %d). Run mjoyd upgradedb.\n", bcVersion, blockchain.BlockChainVersion)
		}
		blockchain.WriteBlockChainVersion(chainDb, blockchain.BlockChainVersion)
	}

	mjoy.blockchain, err = blockchain.NewBlockChain(chainDb, mjoy.chainConfig, mjoy.engine)
	if err != nil {
		return nil, err
	}
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		logger.Warn("Rewinding chain to upgrade configuration", "err", compat)
		mjoy.blockchain.SetHead(compat.RewindTo)
		blockchain.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}
	mjoy.bloomIndexer.Start(mjoy.blockchain)

	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}

	mjoy.txPool = txprocessor.NewTxPool(config.TxPool, mjoy.chainConfig, mjoy.blockchain)

	if mjoy.protocolManager, err = NewProtocolManager(mjoy.chainConfig, config.SyncMode, config.NetworkId, mjoy.eventMux, mjoy.txPool, mjoy.engine, mjoy.blockchain, chainDb); err != nil {
		return nil, err
	}

	//Init miner
	mjoy.blockproducer = blockproducer.New(mjoy,mjoy.interVm,mjoy.chainConfig,mjoy.EventMux() , mjoy.engine)

	mjoy.ApiBackend = &MjoyApiBackend{mjoy}


	fmt.Println("New......Mjoy")
	return mjoy, nil
}

func makeExtraData(extra []byte) []byte {
	return make([]byte , 0)
}

// CreateDB creates the chain database.
func CreateDB(ctx *node.ServiceContext, config *Config, name string) (database.IDatabase, error) {

	db, err := ctx.OpenDatabase(name, config.DatabaseCache, config.DatabaseHandles)
	if err != nil {
		return nil, err
	}
	if db, ok := db.(*database.LDatabase); ok {
		db.Meter("mjoy/db/chaindata/")
	}
	return db, nil
}

// CreateConsensusEngine creates the required type of consensus engine instance for an Mjoy service
func CreateConsensusEngine(mjoy *Mjoy) consensus.Engine {
	engine := consensus.NewBasicEngine(nil)
	return engine
}

func (s *Mjoy) SetEngineKey(pri *ecdsa.PrivateKey) {
	switch v := s.engine.(type) {
	case *consensus.Engine_basic:
		v.SetKey(pri)
	}
}

// APIs returns the collection of RPC services the mjoy package offers.
// NOTE, some of these services probably need to be moved to somewhere else.

func (s *Mjoy) APIs() []rpc.API {
	apis := mjoyapi.GetAPIs(s.ApiBackend)

	//create New
	//apis := make([]rpc.API , 0)

	return append(apis, []rpc.API{
		{
			Namespace: "mjoy",
			Version:   "1.0",
			Service:   NewPublicMjoyAPI(s),
			Public:    true,
		},{
			Namespace:"blockproducer",
			Version:"1.0",
			Service:NewPrivateBlockproducerAPI(s),
			Public:true,
		},
	}...)

}

func (s *Mjoy) ResetWithGenesisBlock(gb *block.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *Mjoy) Coinbase() (eb accounts.Account, err error) {

	if wallets := s.AccountManager().Wallets(); len(wallets) > 0 {
		if accounts := wallets[0].Accounts(); len(accounts) > 0 {
			coinbase := accounts[0].Address

			s.lock.Lock()
			s.coinbase = coinbase
			s.lock.Unlock()

			logger.Infof("Coinbase automatically configured address:0x%x\n" , coinbase)
			return accounts[0], nil
		}
	}
	return accounts.Account{}, fmt.Errorf("Coinbase must be explicitly specified")
}

// set in js console via admin interface or wrapper from cli flags
func (self *Mjoy) SetCoinbase(coinbase types.Address) {
	self.lock.Lock()
	self.coinbase = coinbase
	self.lock.Unlock()

	self.blockproducer.SetCoinbase(coinbase)
}

func (s *Mjoy) StartProducing(local bool, password string) error {
	eb, err := s.Coinbase()
	if err != nil {
		logger.Error("Cannot start producing without coinbase", "err", err)
		return fmt.Errorf("coinbase missing: %v", err)
	}

	//get key
	ks := s.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
	key, err := ks.GetKeyWithPassphrase(eb, password)
	if err != nil {
		logger.Error("Cannot start producing without coinbase, get sign key err ", "err", err)
		return fmt.Errorf("get sign key err: %v", err)
	}
	s.SetEngineKey(key)

	if local {
		// If local (CPU) producing is started, we can disable the transaction rejection
		// mechanism introduced to speed sync times. CPU producing on mainnet is ludicrous
		// so noone will ever hit this path, whereas marking sync done on CPU producing
		// will ensure that private networks work in single blockproducer mode too.
		atomic.StoreUint32(&s.protocolManager.acceptTxs, 1)
	}
	go s.blockproducer.Start(eb.Address)
	return nil
}

func (s *Mjoy) StopProducing()         { s.blockproducer.Stop() }
func (s *Mjoy) IsProducing() bool      { return s.blockproducer.Producing() }
func (s *Mjoy) Blockproducer() *blockproducer.Blockproducer { return s.blockproducer }

func (s *Mjoy) AccountManager() *accounts.Manager  { return s.accountManager }
func (s *Mjoy) BlockChain() *blockchain.BlockChain       { return s.blockchain }
func (s *Mjoy) TxPool() *txprocessor.TxPool               { return s.txPool }
func (s *Mjoy) EventMux() *event.TypeMux           { return s.eventMux }
func (s *Mjoy) Engine() consensus.Engine           { return s.engine }
func (s *Mjoy) ChainDb() database.IDatabase            { return s.chainDb }
func (s *Mjoy) IsListening() bool                  { return true } // Always listening
func (s *Mjoy) MjoyVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *Mjoy) NetVersion() uint64                 { return s.networkId }
func (s *Mjoy) Downloader() *downloader.Downloader { return s.protocolManager.downloader }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *Mjoy) Protocols() []p2p.Protocol {
	if s.lesServer == nil {
		return s.protocolManager.SubProtocols
	}
	return append(s.protocolManager.SubProtocols, s.lesServer.Protocols()...)
}

// Start implements node.Service, starting all internal goroutines needed by the
// Mjoy protocol implementation.
func (s *Mjoy) Start(srvr *p2p.Server) error {
	// Start the bloom bits servicing goroutines
	s.startBloomHandlers()
	
	// Figure out a max peers count based on the server limits
	maxPeers := srvr.MaxPeers
	if s.config.LightServ > 0 {
		maxPeers -= s.config.LightPeers
		if maxPeers < srvr.MaxPeers/2 {
			maxPeers = srvr.MaxPeers / 2
		}
	}
	// Start the networking layer and the light server if requested
	s.protocolManager.Start(maxPeers)
	if s.lesServer != nil {
		s.lesServer.Start(srvr)
	}

	_ , err := s.Coinbase()
	if err != nil{
		fmt.Println("[Warn]No CoinBase Do Not Start Producing Block!!!!!!!!!!!!!!!!!s")
	}
	//when start mjoy service,not start blockproducer,except the cmd order we should start it
	if s.config.StartBlockproducerAtStart{
		fmt.Println("Start Blockproducer At Service Start.......................")
		//s.blockproducer.Start(eb)
	}else {
		fmt.Println("Not Start Blockproducer At Service Start........................")
	}


	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Mjoy protocol.
func (s *Mjoy) Stop() error {
	if s.stopDbUpgrade != nil {
		s.stopDbUpgrade()
	}
	s.bloomIndexer.Close()
	s.blockchain.Stop()
	s.protocolManager.Stop()
	if s.lesServer != nil {
		s.lesServer.Stop()
	}
	s.txPool.Stop()
	s.blockproducer.Stop()
	s.eventMux.Stop()

	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}
