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
// @File: producer.go
// @Date: 2018/05/08 17:23:08
////////////////////////////////////////////////////////////////////////////////

package blockproducer

import (
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"mjoy.io/common"
	"mjoy.io/consensus"
	"mjoy.io/core"
	"mjoy.io/core/state"
	"mjoy.io/common/types"
	"mjoy.io/utils/event"
	"mjoy.io/params"
	"mjoy.io/core/transaction"
	"mjoy.io/core/blockchain/block"
	"mjoy.io/utils/database"
	"mjoy.io/core/stateprocessor"
	"mjoy.io/core/blockchain"
	"gopkg.in/fatih/set.v0"
	"mjoy.io/core/interpreter"
	"mjoy.io/core/sdk"
	"mjoy.io/core/interpreter/balancetransfer"
	"mjoy.io/core/interpreter/intertypes"
	"fmt"
)

const (
	resultQueueSize  = 10
	producingLogAtDepth = 5

	// txChanSize is the size of channel listening to TxPreEvent.
	// The number is referenced from the size of tx pool.
	txChanSize = 4096
	// chainHeadChanSize is the size of channel listening to ChainHeadEvent.
	chainHeadChanSize = 10
	// chainSideChanSize is the size of channel listening to ChainSideEvent.
	chainSideChanSize = 10
)


// Agent can register themself with the worker
type Agent interface {
	Work() chan<- *Work
	SetReturnCh(chan<- *Result)
	Stop()
	Start()
	GetHashRate() int64
}

// Work is the workers current environment and holds
// all of the current state information
type Work struct {
	config *params.ChainConfig

	signer transaction.Signer

	state     *state.StateDB // apply state changes here
	stateRootHash types.Hash
	dbCache   *stateprocessor.DbCache
	ancestors *set.Set       // ancestor set
	family    *set.Set       // family set
	tcount    int            // tx count in cycle
	Block *block.Block // the new block
	header   *block.Header
	txs      []*transaction.Transaction
	receipts []*transaction.Receipt
	mjoy     Backend
	inter    Interpreter
	createdAt time.Time
}

type Result struct {
	Work  *Work
	Block *block.Block
}

// worker is the main object which takes care of applying messages to the new state
type producer struct {
	config *params.ChainConfig
	engine consensus.Engine

	mu sync.Mutex

	// update loop
	mux          *event.TypeMux
	txCh         chan core.TxPreEvent
	txSub        event.Subscription
	chainHeadCh  chan core.ChainHeadEvent
	chainHeadSub event.Subscription
	chainSideCh  chan core.ChainSideEvent
	chainSideSub event.Subscription
	wg           sync.WaitGroup

	agents map[Agent]struct{}
	recv   chan *Result

	mjoy     Backend
	inter    Interpreter
	chain   *blockchain.BlockChain
	proc    blockchain.Validator
	chainDb database.IDatabase

	coinbase types.Address
	//extra    []byte

	currentMu sync.Mutex
	current   *Work

	unconfirmed *unconfirmedBlocks // set of locally produced blocks pending canonicalness confirmations

	// atomic status counters
	producing int32
	atWork int32
}

func newProducer(config *params.ChainConfig, engine consensus.Engine, coinbase types.Address, mjoy Backend,inter Interpreter ,  mux *event.TypeMux) *producer {
	producer := &producer{
		config:         config,
		engine:         engine,
		mjoy:            mjoy,
		inter:          inter,
		mux:            mux,
		txCh:           make(chan core.TxPreEvent, txChanSize),
		chainHeadCh:    make(chan core.ChainHeadEvent, chainHeadChanSize),
		chainSideCh:    make(chan core.ChainSideEvent, chainSideChanSize),
		chainDb:        mjoy.ChainDb(),
		recv:           make(chan *Result, resultQueueSize),
		chain:          mjoy.BlockChain(),
		proc:           mjoy.BlockChain().Validator(),
		coinbase:       coinbase,
		agents:         make(map[Agent]struct{}),
		unconfirmed:    newUnconfirmedBlocks(mjoy.BlockChain(), producingLogAtDepth),
	}
	// Subscribe TxPreEvent for tx pool
	producer.txSub = mjoy.TxPool().SubscribeTxPreEvent(producer.txCh)
	// Subscribe events for blockchain
	producer.chainHeadSub = mjoy.BlockChain().SubscribeChainHeadEvent(producer.chainHeadCh)
	producer.chainSideSub = mjoy.BlockChain().SubscribeChainSideEvent(producer.chainSideCh)
	go producer.update()

	go producer.wait()

	//producer.commitNewWork()

	return producer
}

func (self *producer) setCoinbase(addr types.Address) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.coinbase = addr
}

func (self *producer) setExtra(extra []byte) {
	self.mu.Lock()
	defer self.mu.Unlock()
}

func (self *producer) pending() (*block.Block, *state.StateDB) {
	self.currentMu.Lock()
	defer self.currentMu.Unlock()

	if atomic.LoadInt32(&self.producing) == 0 {
		return block.NewBlock(
			self.current.header,
			self.current.txs,
			self.current.receipts,
		), self.current.state.Copy()
	}
	return self.current.Block, self.current.state.Copy()
}

func (self *producer) pendingBlock() *block.Block {
	self.currentMu.Lock()
	defer self.currentMu.Unlock()

	if atomic.LoadInt32(&self.producing) == 0 {
		return block.NewBlock(
			self.current.header,
			self.current.txs,
			self.current.receipts,
		)
	}
	return self.current.Block
}

func (self *producer) start() {
	self.mu.Lock()
	defer self.mu.Unlock()

	atomic.StoreInt32(&self.producing, 1)

	// spin up agents
	for agent := range self.agents {
		agent.Start()
	}
}

func (self *producer) stop() {
	self.wg.Wait()

	self.mu.Lock()
	defer self.mu.Unlock()
	if atomic.LoadInt32(&self.producing) == 1 {
		for agent := range self.agents {
			agent.Stop()
		}
	}
	atomic.StoreInt32(&self.producing, 0)
	atomic.StoreInt32(&self.atWork, 0)
}

func (self *producer) register(agent Agent) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.agents[agent] = struct{}{}
	agent.SetReturnCh(self.recv)
}

func (self *producer) unregister(agent Agent) {
	self.mu.Lock()
	defer self.mu.Unlock()
	delete(self.agents, agent)
	agent.Stop()
}

func (self *producer) update() {
	defer self.txSub.Unsubscribe()
	defer self.chainHeadSub.Unsubscribe()
	defer self.chainSideSub.Unsubscribe()

	for {
		// A real event arrived, process interesting content
		select {
		// Handle ChainHeadEvent
		case <-self.chainHeadCh:
			self.commitNewWork()

		// Handle TxPreEvent
		case ev := <-self.txCh:
			_ = ev
			// Apply transaction to the pending state if we're not producing
			//if atomic.LoadInt32(&self.producing) == 0 {
			//	self.currentMu.Lock()
			//
			//	acc, _ := transaction.Sender(self.current.signer, ev.Tx)
			//
			//	txs := map[types.Address]transaction.Transactions{acc: {ev.Tx}}
			//	txset := transaction.NewTransactionsForProducing(self.current.signer, txs)
			//
			//	self.current.commitTransactions(self.mux, txset, self.chain, self.coinbase)
			//	self.currentMu.Unlock()
			//} else {
			//	// If we're producing, but nothing is being processed, wake on new transactions
			//	self.commitNewWork()
			//
			//}

		// System stopped
		case <-self.txSub.Err():
			return
		case <-self.chainHeadSub.Err():
			return
		case <-self.chainSideSub.Err():
			return
		}
	}
}

func (self *producer) wait() {
	for {
		mustCommitNewWork := true

		for result := range self.recv {
			atomic.AddInt32(&self.atWork, -1)

			if result == nil {
				continue
			}
			block := result.Block
			work := result.Work

			// Update the block hash in all logs since it is now available and not when the
			// receipt/log of individual transactions were created.

			for _, r := range work.receipts {
				for _, l := range r.Logs {
					l.BlockHash = block.Hash()
				}
			}
			for _, log := range work.state.Logs() {
				log.BlockHash = block.Hash()
			}
			stat, err := self.chain.WriteBlockAndState(block, work.receipts, work.state, work.dbCache)
			if err != nil {
				logger.Error("Failed writing block to chain", "err", err)
				continue
			}

			fmt.Println(block)

			// check if canon block and write transactions
			if stat == blockchain.CanonStatTy {
				// implicit by posting ChainHeadEvent
				mustCommitNewWork = false
			}
			// Broadcast the block and announce chain insertion event
			self.mux.Post(core.NewProducedBlockEvent{Block: block})
			var (
				events []interface{}
				logs   = work.state.Logs()
			)
			events = append(events, core.ChainEvent{Block: block, Hash: block.Hash(), Logs: logs})
			if stat == blockchain.CanonStatTy {
				events = append(events, core.ChainHeadEvent{Block: block})
			}
			self.chain.PostChainEvents(events, logs)

			// Insert the block into the set of pending ones to wait for confirmations
			self.unconfirmed.Insert(block.NumberU64(), block.Hash())

			if mustCommitNewWork {
				self.commitNewWork()
			}
		}
	}
}

// push sends a new work task to currently live blockproducer agents.
func (self *producer) push(work *Work) {
	if atomic.LoadInt32(&self.producing) != 1 {
		return
	}
	for agent := range self.agents {

		atomic.AddInt32(&self.atWork, 1)
		if ch := agent.Work(); ch != nil {
			ch <- work
		}
	}
}
// makeCurrent creates a new environment for the current cycle.
func (self *producer) makeCurrent(parent *block.Block, header *block.Header) error {
	state, err := self.chain.StateAt(parent.Root())
	if err != nil {
		return err
	}
	work := &Work{
		config:    self.config,
		signer:    transaction.NewMSigner(self.config.ChainId),
		state:     state,
		stateRootHash: parent.Root(),
		dbCache:   &stateprocessor.DbCache{make(map[string]interpreter.MemDatabase)},
		ancestors: set.New(),
		family:    set.New(),
		inter:      self.inter,
		header:    header,
		createdAt: time.Now(),
	}

	// when 08 is processed ancestors contain 07 (quick block)
	for _, ancestor := range self.chain.GetBlocksFromHash(parent.Hash(), 7) {
		work.family.Add(ancestor.Hash())
		work.ancestors.Add(ancestor.Hash())
	}

	// Keep track of transactions which return errors so they can be removed
	work.tcount = 0
	self.current = work
	return nil
}

func (self *producer) addTestTransactions(num *big.Int){
	//make action
	//get account 0 = coinbase
	wallets := self.mjoy.AccountManager().Wallets()
	if len(wallets) < 2 {
		return
	}
	w0 := wallets[0]
	w1 := wallets[1]
	//make params
	param := balancetransfer.MakaBalanceTransferParam(w0.Accounts()[0].Address , w1.Accounts()[0].Address , 1000)
	//make action
	action := transaction.MakeAction(balancetransfer.BalanceTransferAddress , param)
	actions := transaction.ActionSlice{}
	actions = append(actions , action)
	//make tx
	nc := self.mjoy.TxPool().State().GetNonce(w0.Accounts()[0].Address)

	tx := transaction.NewTransaction(nc , actions)
	//sign tx
	txSign , err := w0.SignTxWithPassphrase(w0.Accounts()[0] , "123" ,tx , self.config.ChainId)
	if err != nil {
		logger.Errorf("w0.SignTxWithPassphrase err :" , err.Error())
		return
	}

	//add txSign to txpool
	self.mjoy.TxPool().AddRemote(txSign)


}

func (self *producer) commitNewWork() {

	self.mu.Lock()
	defer self.mu.Unlock()
	self.currentMu.Lock()
	defer self.currentMu.Unlock()

	tstart := time.Now()
	parent := self.chain.CurrentBlock()

	tstamp := tstart.Unix()
	if parent.Time().Cmp(new(big.Int).SetInt64(tstamp)) >= 0 {
		tstamp = parent.Time().Int64() + 1
	}

	// this will ensure we're not going off too far in the future
	if now := time.Now().Unix(); tstamp > now+1 {
		wait := time.Duration(tstamp-now) * time.Second
		logger.Info("Producing too far in the future", "wait", common.PrettyDuration(wait))
		time.Sleep(wait)
	}

	num := parent.Number()

	header := &block.Header{
		ParentHash: parent.Hash(),
		Number:     &types.BigInt{*num.Add(num, common.Big1)},
		Time:       &types.BigInt{*big.NewInt(tstamp)},
	}
	// Only set the coinbase if we are producing (avoid spurious block rewards)
	if atomic.LoadInt32(&self.producing) == 1 {
		header.BlockProducer = self.coinbase
	}

	if err := self.engine.Prepare(self.chain, header); err != nil {
		logger.Error("Failed to prepare header for producing", "err", err)
		return
	}


	// Could potentially happen if starting to produce block in an odd state.
	err := self.makeCurrent(parent, header)
	if err != nil {
		logger.Error("Failed to create producing context", "err", err)
		return
	}
	// Create the current work task and check any fork transitions needed
	work := self.current

	//add test tx
	//self.addTestTransactions(num)
	pending, err := self.mjoy.TxPool().Pending()
	if err != nil {
		logger.Error("Failed to fetch pending transactions", "err", err)
		return
	}
	logger.Info(">>>>>>>>>PendingTx Len:" , len(pending))


	//txs := transaction.NewTransactionsForProducing(self.current.signer, pending)
	actions := transaction.ActionSlice{}
	action := transaction.Action{&types.Address{},balancetransfer.MakeActionParamsReword(header.BlockProducer)}
	actions = append(actions, action)

	tx := transaction.NewTransaction(num.Uint64() - 1 , actions)

	txReword, err := transaction.SignTx(tx,self.current.signer, params.RewordPrikey)
	if err != nil {
		logger.Error("Failed to make reword transaction", "err", err)
		return
	}
	txReword.Priority = big.NewInt(10)
	txs := transaction.NewTransactionsByPriorityAndNonce(self.current.signer , pending, txReword)

	sdkHandler := sdk.NewTmpStatusManager(self.chain.GetDb(), work.state,self.coinbase)
	vmHandler := interpreter.NewVm()
	sysparam := intertypes.MakeSystemParams(sdkHandler,vmHandler )
	work.commitTransactions(self.mux, txs, self.chain, self.coinbase , sysparam)


	// Create the new block to seal with the consensus engine
	if work.Block, err = self.engine.Finalize(self.chain, header, work.state, work.txs, work.receipts, true); err != nil {
		logger.Error("Failed to finalize block for sealing", "err", err)
		return
	}
	// We only care about logging if we're actually producing.
	if atomic.LoadInt32(&self.producing) == 1 {
		logger.Info("Commit new producing work", "number", work.Block.Number(), "txs", work.tcount, "elapsed", common.PrettyDuration(time.Since(tstart)))
		self.unconfirmed.Shift(work.Block.NumberU64() - 1)
	}
	work.mjoy = self.mjoy
	self.push(work)
}

func (env *Work) commitTransactions(mux *event.TypeMux, txs *transaction.TransactionsByPriorityAndNonce, bc *blockchain.BlockChain, coinbase types.Address , sysparam *intertypes.SystemParams) {

	var coalescedLogs []*transaction.Log

	dbcache := env.dbCache

	for {

		// Retrieve the next transaction and abort if all done
		tx := txs.Peek()
		if tx == nil {
			break
		}
		// Error may be ignored here. The error has already been checked
		// during transaction acceptance is the transaction pool.
		//
		// We use the eip155 signer regardless of the current hf.
		from, _ := transaction.Sender(env.signer, tx)
		// Check whether the tx is replay protected. If we're not in the EIP155 hf
		// phase, start ignoring the sender until we do.
		if false{
			if tx.Protected(){
				logger.Tracef("Ignoring reply protected transaction hash:%x\n" , tx.Hash())

				txs.Pop()
				continue
			}
		}
		//err := env.inter.SendWork(from , tx.Data.Actions)
		//if err != nil{
		//	txs.Shift()
		//}else{
		//	txs.Shift()
		//}

		// Start executing the transaction
		env.state.Prepare(tx.Hash(), types.Hash{}, env.tcount)

		err, logs := env.commitTransaction(tx, bc, coinbase, dbcache , sysparam)
		switch err {
		case core.ErrNonceTooLow:
			// New head notification data race between the transaction pool and blockproducer, shift
			logger.Info("Skipping transaction with low nonce", "sender", from, "nonce", tx.Nonce())
			txs.Shift()

		case core.ErrNonceTooHigh:
			// Reorg notification data race between the transaction pool and blockproducer, skip account =
			logger.Info("Skipping account with hight nonce", "sender", from, "nonce", tx.Nonce())
			txs.Pop()

		case nil:
			// Everything ok, collect the logs and shift in the next transaction from the same account
			coalescedLogs = append(coalescedLogs, logs...)
			env.tcount++
			txs.Shift()

		default:
			// Strange error, discard the transaction and get the next in line (note, the
			// nonce-too-high clause will prevent us from executing in vain).
			logger.Debug("Transaction failed, account skipped", "hash", tx.Hash().String(), "err", err)
			txs.Shift()
		}
	}

	if len(coalescedLogs) > 0 || env.tcount > 0 {
		// make a copy, the state caches the logs and these logs get "upgraded" from pending to produced block
		// logs by filling in the block hash when the block was produced block by the local blockproducer. This can
		// cause a race condition if a log was "upgraded" before the PendingLogsEvent is processed.
		cpy := make([]*transaction.Log, len(coalescedLogs))
		for i, l := range coalescedLogs {
			cpy[i] = new(transaction.Log)
			*cpy[i] = *l
		}
		go func(logs []*transaction.Log, tcount int) {
			if len(logs) > 0 {
				mux.Post(core.PendingLogsEvent{Logs: logs})
			}
			if tcount > 0 {
				mux.Post(core.PendingStateEvent{})
			}
		}(cpy, env.tcount)
	}
}

func (env *Work) commitTransaction(tx *transaction.Transaction, bc *blockchain.BlockChain, coinbase types.Address, cache *stateprocessor.DbCache , sysparam *intertypes.SystemParams) (error, []*transaction.Log) {
	snap := env.state.Snapshot()
	//                                ApplyTransaction(this.config,&coinbase,this.state ,header,tx)
	receipt, err := stateprocessor.ApplyTransaction(env.config, &coinbase,  env.state, env.header, tx, cache , sysparam)
	if err != nil {
		env.state.RevertToSnapshot(snap)
		return err, nil
	}
	env.txs = append(env.txs, tx)
	env.receipts = append(env.receipts, receipt)

	return nil, receipt.Logs
}
