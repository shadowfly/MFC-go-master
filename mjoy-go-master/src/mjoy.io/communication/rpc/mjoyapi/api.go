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
// @File: api.go
// @Date: 2018/05/08 15:18:08
////////////////////////////////////////////////////////////////////////////////

package mjoyapi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"mjoy.io/accounts"
	"mjoy.io/accounts/keystore"
	"mjoy.io/common/types/util/hex"
	"mjoy.io/core/transaction"
	"mjoy.io/common/types"
	"mjoy.io/utils/crypto"
	"mjoy.io/common/math"
	"mjoy.io/communication/rpc"
	"mjoy.io/core/blockchain/block"
	"github.com/tinylib/msgp/msgp"
	"mjoy.io/core/blockchain"
	"mjoy.io/core/sdk"
	"mjoy.io/core/interpreter"
	"mjoy.io/core/interpreter/intertypes"
)


// PublicMjoyAPI provides an API to access mjoy related information.
// It offers only methods that operate on public data that is freely available to anyone.
type PublicMjoyAPI struct {
	b Backend
}

// NewPublicMjoyAPI creates a new mjoy protocol API.
func NewPublicMjoyAPI(b Backend) *PublicMjoyAPI {
	return &PublicMjoyAPI{b}
}


// ProtocolVersion returns the current mjoy protocol version this node supports
func (s *PublicMjoyAPI) ProtocolVersion() hex.Uint {
	return hex.Uint(s.b.ProtocolVersion())
}

// Syncing returns false in case the node is currently not syncing with the network. It can be up to date or has not
// yet received the latest block headers from its pears. In case it is synchronizing:
// - startingBlock: block number this node started to synchronise from
// - currentBlock:  block number this node is currently importing
// - highestBlock:  block number of the highest block header this node has received from peers
// - pulledStates:  number of state entries processed until now
// - knownStates:   number of known state entries that still need to be pulled
func (s *PublicMjoyAPI) Syncing() (interface{}, error) {
	progress := s.b.Downloader().Progress()

	// Return not syncing if the synchronisation already completed
	if progress.CurrentBlock >= progress.HighestBlock {
		return false, nil
	}
	// Otherwise gather the block sync stats
	return map[string]interface{}{
		"startingBlock": hex.Uint64(progress.StartingBlock),
		"currentBlock":  hex.Uint64(progress.CurrentBlock),
		"highestBlock":  hex.Uint64(progress.HighestBlock),
		"pulledStates":  hex.Uint64(progress.PulledStates),
		"knownStates":   hex.Uint64(progress.KnownStates),
	}, nil
}

// PublicTxPoolAPI offers and API for the transaction pool. It only operates on data that is non confidential.
type PublicTxPoolAPI struct {
	b Backend
}

// NewPublicTxPoolAPI creates a new tx pool service that gives information about the transaction pool.
func NewPublicTxPoolAPI(b Backend) *PublicTxPoolAPI {
	return &PublicTxPoolAPI{b}
}

// Content returns the transactions contained within the transaction pool.
func (s *PublicTxPoolAPI) Content() map[string]map[string]map[string]*RPCTransaction {
	content := map[string]map[string]map[string]*RPCTransaction{
		"pending": make(map[string]map[string]*RPCTransaction),
		"queued":  make(map[string]map[string]*RPCTransaction),
	}
	pending, queue := s.b.TxPoolContent()

	// Flatten the pending transactions
	for account, txs := range pending {
		dump := make(map[string]*RPCTransaction)
		for _, tx := range txs {
			dump[fmt.Sprintf("%d", tx.Nonce())] = newRPCPendingTransaction(tx)
		}
		content["pending"][account.Hex()] = dump
	}
	// Flatten the queued transactions
	for account, txs := range queue {
		dump := make(map[string]*RPCTransaction)
		for _, tx := range txs {
			dump[fmt.Sprintf("%d", tx.Nonce())] = newRPCPendingTransaction(tx)
		}
		content["queued"][account.Hex()] = dump
	}
	return content
}

// Status returns the number of pending and queued transaction in the pool.
func (s *PublicTxPoolAPI) Status() map[string]hex.Uint {
	pending, queue := s.b.Stats()
	return map[string]hex.Uint{
		"pending": hex.Uint(pending),
		"queued":  hex.Uint(queue),
	}
}

// Inspect retrieves the content of the transaction pool and flattens it into an
// easily inspectable list.
func (s *PublicTxPoolAPI) Inspect() map[string]map[string]map[string]string {
	content := map[string]map[string]map[string]string{
		"pending": make(map[string]map[string]string),
		"queued":  make(map[string]map[string]string),
	}
	pending, queue := s.b.TxPoolContent()

	// Define a formatter to flatten a transaction into a string
	var format = func(tx *transaction.Transaction) string {
		//if to := tx.To(); to != nil {
			//return fmt.Sprintf("%s: %v wei", tx.To().Hex(), tx.Value())
			//return "for test"
		//}
		//return fmt.Sprintf("contract creation: %v wei", tx.Value())
		return "for test"
	}

	// Flatten the pending transactions
	for account, txs := range pending {
		dump := make(map[string]string)
		for _, tx := range txs {
			dump[fmt.Sprintf("%d", tx.Nonce())] = format(tx)
		}
		content["pending"][account.Hex()] = dump
	}
	// Flatten the queued transactions
	for account, txs := range queue {
		dump := make(map[string]string)
		for _, tx := range txs {
			dump[fmt.Sprintf("%d", tx.Nonce())] = format(tx)
		}
		content["queued"][account.Hex()] = dump
	}
	return content
}

// PublicAccountAPI provides an API to access accounts managed by this node.
// It offers only methods that can retrieve accounts.
type PublicAccountAPI struct {
	am *accounts.Manager
}

// NewPublicAccountAPI creates a new PublicAccountAPI.
func NewPublicAccountAPI(am *accounts.Manager) *PublicAccountAPI {
	return &PublicAccountAPI{am: am}
}

// Accounts returns the collection of accounts this node manages
func (s *PublicAccountAPI) Accounts() []types.Address {
	addresses := make([]types.Address, 0) // return [] instead of nil if empty
	for _, wallet := range s.am.Wallets() {
		for _, account := range wallet.Accounts() {
			addresses = append(addresses, account.Address)
		}
	}
	return addresses
}

// PrivateAccountAPI provides an API to access accounts managed by this node.
// It offers methods to create, (un)lock en list accounts. Some methods accept
// passwords and are therefore considered private by default.
type PrivateAccountAPI struct {
	am        *accounts.Manager
	nonceLock *AddrLocker
	b         Backend
}

// NewPrivateAccountAPI create a new PrivateAccountAPI.
func NewPrivateAccountAPI(b Backend, nonceLock *AddrLocker) *PrivateAccountAPI {
	return &PrivateAccountAPI{
		am:        b.AccountManager(),
		nonceLock: nonceLock,
		b:         b,
	}
}

// ListAccounts will return a list of addresses for accounts this node manages.
func (s *PrivateAccountAPI) ListAccounts() []types.Address {
	addresses := make([]types.Address, 0) // return [] instead of nil if empty
	for _, wallet := range s.am.Wallets() {
		for _, account := range wallet.Accounts() {
			addresses = append(addresses, account.Address)
		}
	}
	return addresses
}

// rawWallet is a JSON representation of an accounts.Wallet interface, with its
// data contents extracted into plain fields.
type rawWallet struct {
	URL      string             `json:"url"`
	Status   string             `json:"status"`
	Failure  string             `json:"failure,omitempty"`
	Accounts []accounts.Account `json:"accounts,omitempty"`
}

// ListWallets will return a list of wallets this node manages.
func (s *PrivateAccountAPI) ListWallets() []rawWallet {
	wallets := make([]rawWallet, 0) // return [] instead of nil if empty
	for _, wallet := range s.am.Wallets() {
		status, failure := wallet.Status()

		raw := rawWallet{
			URL:      wallet.URL().String(),
			Status:   status,
			Accounts: wallet.Accounts(),
		}
		if failure != nil {
			raw.Failure = failure.Error()
		}
		wallets = append(wallets, raw)
	}
	return wallets
}

// OpenWallet initiates a hardware wallet opening procedure, establishing a USB
// connection and attempting to authenticate via the provided passphrase. Note,
// the method may return an extra challenge requiring a second open (e.g. the
// Trezor PIN matrix challenge).
func (s *PrivateAccountAPI) OpenWallet(url string, passphrase *string) error {
	wallet, err := s.am.Wallet(url)
	if err != nil {
		return err
	}
	pass := ""
	if passphrase != nil {
		pass = *passphrase
	}
	return wallet.Open(pass)
}

// DeriveAccount requests a HD wallet to derive a new account, optionally pinning
// it for later reuse.
func (s *PrivateAccountAPI) DeriveAccount(url string, path string, pin *bool) (accounts.Account, error) {
	wallet, err := s.am.Wallet(url)
	if err != nil {
		return accounts.Account{}, err
	}
	derivPath, err := accounts.ParseDerivationPath(path)
	if err != nil {
		return accounts.Account{}, err
	}
	if pin == nil {
		pin = new(bool)
	}
	return wallet.Derive(derivPath, *pin)
}

// NewAccount will create a new account and returns the address for the new account.
func (s *PrivateAccountAPI) NewAccount(password string) (types.Address, error) {
	acc, err := fetchKeystore(s.am).NewAccount(password)
	if err == nil {
		return acc.Address, nil
	}
	return types.Address{}, err
}

// fetchKeystore retrives the encrypted keystore from the account manager.
func fetchKeystore(am *accounts.Manager) *keystore.KeyStore {
	return am.Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
}

// ImportRawKey stores the given hex encoded ECDSA key into the key directory,
// encrypting it with the passphrase.
func (s *PrivateAccountAPI) ImportRawKey(privkey string, password string) (types.Address, error) {
	key, err := crypto.HexToECDSA(privkey)
	if err != nil {
		return types.Address{}, err
	}
	acc, err := fetchKeystore(s.am).ImportECDSA(key, password)
	return acc.Address, err
}

// UnlockAccount will unlock the account associated with the given address with
// the given password for duration seconds. If duration is nil it will use a
// default of 300 seconds. It returns an indication if the account was unlocked.
func (s *PrivateAccountAPI) UnlockAccount(addr types.Address, password string, duration *uint64) (bool, error) {
	const max = uint64(time.Duration(math.MaxInt64) / time.Second)
	var d time.Duration
	if duration == nil {
		d = 300 * time.Second
	} else if *duration > max {
		return false, errors.New("unlock duration too large")
	} else {
		d = time.Duration(*duration) * time.Second
	}
	err := fetchKeystore(s.am).TimedUnlock(accounts.Account{Address: addr}, password, d)
	return err == nil, err
}

// LockAccount will lock the account associated with the given address when it's unlocked.
func (s *PrivateAccountAPI) LockAccount(addr types.Address) bool {
	return fetchKeystore(s.am).Lock(addr) == nil
}

// SendTransaction will create a transaction from the given arguments and
// tries to sign it with the key associated with args.To. If the given passwd isn't
// able to decrypt the key it fails.
func (s *PrivateAccountAPI) SendTransaction(ctx context.Context, args SendTxArgs, passwd string) (types.Hash, error) {
	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: args.From}

	wallet, err := s.am.Find(account)
	if err != nil {
		return types.Hash{}, err
	}

	if args.Nonce == nil {
		// Hold the addresse's mutex around signing to prevent concurrent assignment of
		// the same nonce to multiple accounts.
		s.nonceLock.LockAddr(args.From)
		defer s.nonceLock.UnlockAddr(args.From)
	}

	// Set some sanity defaults and terminate on failure
	if err := args.setDefaults(ctx, s.b); err != nil {
		return types.Hash{}, err
	}
	// Assemble the transaction and sign with the wallet
	tx := args.toTransaction()

	chainID := s.b.ChainConfig().ChainId
	signed, err := wallet.SignTxWithPassphrase(account, passwd, tx, chainID)
	if err != nil {
		return types.Hash{}, err
	}
	return submitTransaction(ctx, s.b, signed)
}

// signHash is a helper function that calculates a hash for the given message that can be
// safely used to calculate a signature from.
//
// The hash is calulcated as
//   keccak256("\x19Mjoy Signed Message:\n"${message length}${message}).
//
// This gives context to the signed message and prevents signing of transactions.
func signHash(data []byte) []byte {
	msg := fmt.Sprintf("\x19Mjoy Signed Message:\n%d%s", len(data), data)
	return crypto.Keccak256([]byte(msg))
}

// Sign calculates an mjoy ECDSA signature for:
// keccack256("\x19mjoy Signed Message:\n" + len(message) + message))
//
// Note, the produced signature conforms to the secp256k1 curve R, S and V values,
// where the V value will be 27 or 28 for legacy reasons.
//
// The key used to calculate the signature is decrypted with the given password.
//
func (s *PrivateAccountAPI) Sign(ctx context.Context, data hex.Bytes, addr types.Address, passwd string) (hex.Bytes, error) {
	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: addr}

	wallet, err := s.b.AccountManager().Find(account)
	if err != nil {
		return nil, err
	}
	// Assemble sign the data with the wallet
	signature, err := wallet.SignHashWithPassphrase(account, passwd, signHash(data))
	if err != nil {
		return nil, err
	}
	signature[64] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper
	return signature, nil
}

// EcRecover returns the address for the account that was used to create the signature.
// Note, this function is compatible with mjoy_sign and personal_sign. As such it recovers
// the address of:
// hash = keccak256("\x19Mjoy Signed Message:\n"${message length}${message})
// addr = ecrecover(hash, signature)
//
// Note, the signature must conform to the secp256k1 curve R, S and V values, where
// the V value must be be 27 or 28 for legacy reasons.
//
func (s *PrivateAccountAPI) EcRecover(ctx context.Context, data, sig hex.Bytes) (types.Address, error) {
	if len(sig) != 65 {
		return types.Address{}, fmt.Errorf("signature must be 65 bytes long")
	}
	if sig[64] != 27 && sig[64] != 28 {
		return types.Address{}, fmt.Errorf("invalid mjoy signature (V is not 27 or 28)")
	}
	sig[64] -= 27 // Transform yellow paper V from 27/28 to 0/1

	rpk, err := crypto.Ecrecover(signHash(data), sig)
	if err != nil {
		return types.Address{}, err
	}
	pubKey := crypto.ToECDSAPub(rpk)
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	return recoveredAddr, nil
}


// PublicBlockChainAPI provides an API to access the mjoy blockchain.
// It offers only methods that operate on public data that is freely available to anyone.
type PublicBlockChainAPI struct {
	b Backend
}

// NewPublicBlockChainAPI creates a new mjoy blockchain API.
func NewPublicBlockChainAPI(b Backend) *PublicBlockChainAPI {
	return &PublicBlockChainAPI{b}
}

// BlockNumber returns the block number of the chain head.
func (s *PublicBlockChainAPI) BlockNumber() *big.Int {
	header, _ := s.b.HeaderByNumber(context.Background(), rpc.LatestBlockNumber) // latest header should always be available
	return &header.Number.IntVal
}

// BlockNumber returns the block number of the chain head.
func (s *PublicBlockChainAPI) GetNum() int {
	return 20
}

// GetStoragePara returns the storage parameters of certain contract
// The detail parameter is defined by contact interpreter
func (s *PublicBlockChainAPI) GetStorageParameter(ctx context.Context, actionArg SendTxAction,blockNr rpc.BlockNumber) (hex.Bytes, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, err
	}
	fmt.Println("=====================================>")
	sdkHandler := sdk.NewTmpStatusManager(s.b.ChainDb(), state, types.Address{})
	vmHandler := interpreter.NewVm()
	sysparam := intertypes.MakeSystemParams(sdkHandler, vmHandler)
	//package param
	action := transaction.Action{}
	action.Address = actionArg.Address
	action.Params = *actionArg.Params

	getResult := vmHandler.GetStorage(types.Address{} , action, sysparam)
	if getResult.Err != nil {
		return nil , getResult.Err
	}

	fmt.Println("getResult:" , string(getResult.Var))

	hexbyte := make(hex.Bytes, len(getResult.Var))
	copy(hexbyte, getResult.Var)

	return hexbyte , nil

	//b := state.GetBalance(address)
	//return nil, state.Error()
}

// GetBlockByNumber returns the requested block. When blockNr is -1 the chain head is returned. When fullTx is true all
// transactions in the block are returned in full detail, otherwise only the transaction hash is returned.
func (s *PublicBlockChainAPI) GetBlockByNumber(ctx context.Context, blockNr rpc.BlockNumber, fullTx bool) (map[string]interface{}, error) {
	block, err := s.b.BlockByNumber(ctx, blockNr)
	if block != nil {
		response, err := s.rpcOutputBlock(block, true, fullTx)
		if err == nil && blockNr == rpc.PendingBlockNumber {
			// Pending blocks need to nil out a few fields
			for _, field := range []string{"hash", "nonce", "blockproducer"} {
				response[field] = nil
			}
		}
		return response, err
	}
	return nil, err
}

// GetBlockByHash returns the requested block. When fullTx is true all transactions in the block are returned in full
// detail, otherwise only the transaction hash is returned.
func (s *PublicBlockChainAPI) GetBlockByHash(ctx context.Context, blockHash types.Hash, fullTx bool) (map[string]interface{}, error) {
	block, err := s.b.GetBlock(ctx, blockHash)
	if block != nil {
		return s.rpcOutputBlock(block, true, fullTx)
	}
	return nil, err
}

// GetCode returns the code stored at the given address in the state for the given block number.
func (s *PublicBlockChainAPI) GetCode(ctx context.Context, address types.Address, blockNr rpc.BlockNumber) (hex.Bytes, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, err
	}
	code := state.GetCode(address)
	return code, state.Error()
}

// GetStorageAt returns the storage from the state at the given address, key and
// block number. The rpc.LatestBlockNumber and rpc.PendingBlockNumber meta block
// numbers are also allowed.
func (s *PublicBlockChainAPI) GetStorageAt(ctx context.Context, address types.Address, key string, blockNr rpc.BlockNumber) (hex.Bytes, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, err
	}
	res := state.GetState(address, types.HexToHash(key))
	return res[:], state.Error()
}

type ConsensusDataHex struct {
	Id      string         `json:"id"`
	Para    *hex.Bytes     `json:"data"`
}

// rpcOutputBlock converts the given block to the RPC output which depends on fullTx. If inclTx is true transactions are
// returned. When fullTx is true the returned block contains full transaction details, otherwise it will only contain
// transaction hashes.
func (s *PublicBlockChainAPI) rpcOutputBlock(b *block.Block, inclTx bool, fullTx bool) (map[string]interface{}, error) {
	head := b.Header() // copies the header once

	hexbyte := make(hex.Bytes, len(head.ConsensusData.Para))
	copy(hexbyte, head.ConsensusData.Para)
	cHex := ConsensusDataHex{head.ConsensusData.Id, &hexbyte}
	fields := map[string]interface{}{
		"number":           	(*hex.Big)(&head.Number.IntVal),
		"hash":             	b.Hash(),
		"parentHash":       	head.ParentHash,
		"logsBloom":        	head.Bloom,
		"stateRoot":        	head.StateRootHash,
		"blockproducer":    	head.BlockProducer,
		"consensusData":     	cHex,
		"size":             	hex.Uint64(uint64(b.Size())),
		"timestamp":        	(*hex.Big)(&head.Time.IntVal),
		"transactionsRoot": 	head.TxRootHash,
		"receiptsRoot":     	head.ReceiptRootHash,
		"R":                	(*hex.Big)(&head.R.IntVal),
		"S":                	(*hex.Big)(&head.S.IntVal),
		"V":                	(*hex.Big)(&head.V.IntVal),
	}

	if inclTx {
		formatTx := func(tx *transaction.Transaction) (interface{}, error) {
			return tx.Hash(), nil
		}

		if fullTx {
			formatTx = func(tx *transaction.Transaction) (interface{}, error) {
				return newRPCTransactionFromBlockHash(b, tx.Hash()), nil
			}
		}

		txs := b.Transactions()
		transactions := make([]interface{}, len(txs))
		var err error
		for i, tx := range b.Transactions() {
			if transactions[i], err = formatTx(tx); err != nil {
				return nil, err
			}
		}
		fields["transactions"] = transactions
	}


	return fields, nil
}

//api test
var rateFlag uint64 = 1

func (s *PublicBlockChainAPI)Forking(ctx context.Context , rate uint64)(uint64){
	rateFlag = rate
	rate = rate + 1
	return rate
}








// RPCTransaction represents a transaction that will serialize to the RPC representation of a transaction
type RPCTransaction struct {
	BlockHash        types.Hash     		 	`json:"blockHash"`
	BlockNumber      *hex.Big       			`json:"blockNumber"`
	From             types.Address  			`json:"from"`
	Hash             types.Hash     			`json:"hash"`
	Nonce            hex.Uint64     			`json:"nonce"`
	TransactionIndex hex.Uint       			`json:"transactionIndex"`
	Actions          []*SendTxAction			`json:"actions"`
	V                *hex.Big       			`json:"v"`
	R                *hex.Big       			`json:"r"`
	S                *hex.Big       			`json:"s"`
}

// newRPCTransaction returns a transaction that will serialize to the RPC
// representation, with the given location metadata set (if available).
func newRPCTransaction(tx *transaction.Transaction, blockHash types.Hash, blockNumber uint64, index uint64) *RPCTransaction {
	var signer transaction.Signer = transaction.NewMSigner(tx.ChainId())

	from, _ := transaction.Sender(signer, tx)
	v, r, s := tx.RawSignatureValues()

	actions := []*SendTxAction{}

	for _, action := range tx.Data.Actions {
		hexbyte := make(hex.Bytes, len(action.Params))
		copy(hexbyte, action.Params)
		actionSend := &SendTxAction{action.Address, &hexbyte}
		actions = append(actions, actionSend)
	}

	result := &RPCTransaction{
		From:     from,
		Hash:     tx.Hash(),
		Nonce:    hex.Uint64(tx.Nonce()),
		V:        (*hex.Big)(v),
		R:        (*hex.Big)(r),
		S:        (*hex.Big)(s),
		Actions:  actions,
	}
	if blockHash != (types.Hash{}) {
		result.BlockHash = blockHash
		result.BlockNumber = (*hex.Big)(new(big.Int).SetUint64(blockNumber))
		result.TransactionIndex = hex.Uint(index)
	}
	return result

	return nil
}

// newRPCPendingTransaction returns a pending transaction that will serialize to the RPC representation
func newRPCPendingTransaction(tx *transaction.Transaction) *RPCTransaction {
	return newRPCTransaction(tx, types.Hash{}, 0, 0)
}

// newRPCTransactionFromBlockIndex returns a transaction that will serialize to the RPC representation.
func newRPCTransactionFromBlockIndex(b *block.Block, index uint64) *RPCTransaction {
	txs := b.Transactions()
	if index >= uint64(len(txs)) {
		return nil
	}
	return newRPCTransaction(txs[index], b.Hash(), b.NumberU64(), index)
}

// newRPCRawTransactionFromBlockIndex returns the bytes of a transaction given a block and a transaction index.
func newRPCRawTransactionFromBlockIndex(b *block.Block, index uint64) hex.Bytes {
	txs := b.Transactions()
	if index >= uint64(len(txs)) {
		return nil
	}
	var buf bytes.Buffer
	msgp.Encode(&buf, txs[index])

	return buf.Bytes()
}

// newRPCTransactionFromBlockHash returns a transaction that will serialize to the RPC representation.
func newRPCTransactionFromBlockHash(b *block.Block, hash types.Hash) *RPCTransaction {
	for idx, tx := range b.Transactions() {
		if tx.Hash() == hash {
			return newRPCTransactionFromBlockIndex(b, uint64(idx))
		}
	}
	return nil
}

// PublicTransactionPoolAPI exposes methods for the RPC interface
type PublicTransactionPoolAPI struct {
	b         Backend
	nonceLock *AddrLocker
}

// NewPublicTransactionPoolAPI creates a new RPC service with methods specific for the transaction pool.
func NewPublicTransactionPoolAPI(b Backend, nonceLock *AddrLocker) *PublicTransactionPoolAPI {
	return &PublicTransactionPoolAPI{b, nonceLock}
}

// GetBlockTransactionCountByNumber returns the number of transactions in the block with the given block number.
func (s *PublicTransactionPoolAPI) GetBlockTransactionCountByNumber(ctx context.Context, blockNr rpc.BlockNumber) *hex.Uint {
	if block, _ := s.b.BlockByNumber(ctx, blockNr); block != nil {
		n := hex.Uint(len(block.Transactions()))
		return &n
	}
	return nil
}

// GetBlockTransactionCountByHash returns the number of transactions in the block with the given hash.
func (s *PublicTransactionPoolAPI) GetBlockTransactionCountByHash(ctx context.Context, blockHash types.Hash) *hex.Uint {
	if block, _ := s.b.GetBlock(ctx, blockHash); block != nil {
		n := hex.Uint(len(block.Transactions()))
		return &n
	}
	return nil
}

// GetTransactionByBlockNumberAndIndex returns the transaction for the given block number and index.
func (s *PublicTransactionPoolAPI) GetTransactionByBlockNumberAndIndex(ctx context.Context, blockNr rpc.BlockNumber, index hex.Uint) *RPCTransaction {
	if block, _ := s.b.BlockByNumber(ctx, blockNr); block != nil {
		return newRPCTransactionFromBlockIndex(block, uint64(index))
	}
	return nil
}

// GetTransactionByBlockHashAndIndex returns the transaction for the given block hash and index.
func (s *PublicTransactionPoolAPI) GetTransactionByBlockHashAndIndex(ctx context.Context, blockHash types.Hash, index hex.Uint) *RPCTransaction {
	if block, _ := s.b.GetBlock(ctx, blockHash); block != nil {
		return newRPCTransactionFromBlockIndex(block, uint64(index))
	}
	return nil
}

// GetRawTransactionByBlockNumberAndIndex returns the bytes of the transaction for the given block number and index.
func (s *PublicTransactionPoolAPI) GetRawTransactionByBlockNumberAndIndex(ctx context.Context, blockNr rpc.BlockNumber, index hex.Uint) hex.Bytes {
	if block, _ := s.b.BlockByNumber(ctx, blockNr); block != nil {
		return newRPCRawTransactionFromBlockIndex(block, uint64(index))
	}
	return nil
}

// GetRawTransactionByBlockHashAndIndex returns the bytes of the transaction for the given block hash and index.
func (s *PublicTransactionPoolAPI) GetRawTransactionByBlockHashAndIndex(ctx context.Context, blockHash types.Hash, index hex.Uint) hex.Bytes {
	if block, _ := s.b.GetBlock(ctx, blockHash); block != nil {
		return newRPCRawTransactionFromBlockIndex(block, uint64(index))
	}
	return nil
}

// GetTransactionCount returns the number of transactions the given address has sent for the given block number
func (s *PublicTransactionPoolAPI) GetTransactionCount(ctx context.Context, address types.Address, blockNr rpc.BlockNumber) (*hex.Uint64, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, err
	}
	nonce := state.GetNonce(address)
	return (*hex.Uint64)(&nonce), state.Error()
}

// GetTransactionByHash returns the transaction for the given hash
func (s *PublicTransactionPoolAPI) GetTransactionByHash(ctx context.Context, hash types.Hash) *RPCTransaction {
	// Try to return an already finalized transaction
	if tx, blockHash, blockNumber, index := blockchain.GetTransaction(s.b.ChainDb(), hash); tx != nil {
		return newRPCTransaction(tx, blockHash, blockNumber, index)
	}
	// No finalized transaction, try to retrieve it from the pool
	if tx := s.b.GetPoolTransaction(hash); tx != nil {
		return newRPCPendingTransaction(tx)
	}
	// Transaction unknown, return as such
	return nil
}

// GetRawTransactionByHash returns the bytes of the transaction for the given hash.
func (s *PublicTransactionPoolAPI) GetRawTransactionByHash(ctx context.Context, hash types.Hash) (hex.Bytes, error) {
	var tx *transaction.Transaction

	// Retrieve a finalized transaction, or a pooled otherwise
	if tx, _, _, _ = blockchain.GetTransaction(s.b.ChainDb(), hash); tx == nil {
		if tx = s.b.GetPoolTransaction(hash); tx == nil {
			// Transaction not found anywhere, abort
			return nil, nil
		}
	}
	// Serialize to MSGP and return
	var buf bytes.Buffer
	err := msgp.Encode(&buf, tx)
	return buf.Bytes(), err
}

// GetTransactionReceipt returns the transaction receipt for the given transaction hash.
func (s *PublicTransactionPoolAPI) GetTransactionReceipt(hash types.Hash) (map[string]interface{}, error) {
	//tx, blockHash, blockNumber, index := blockchain.GetTransaction(s.b.ChainDb(), hash)
	//if tx == nil {
	//	return nil, errors.New("unknown transaction")
	//}
	//receipt, _, _, _ := blockchain.GetReceipt(s.b.ChainDb(), hash) // Old receipts don't have the lookup data available
	//if receipt == nil {
	//	return nil, errors.New("unknown receipt")
	//}
	//
	//var signer transaction.Signer = transaction.NewMSigner(tx.ChainId())
	//
	//from, _ := transaction.Sender(signer, tx)
	//
	//fields := map[string]interface{}{
	//	"blockHash":         blockHash,
	//	"blockNumber":       hex.Uint64(blockNumber),
	//	"transactionHash":   hash,
	//	"transactionIndex":  hex.Uint64(index),
	//	"from":              from,
	//	"to":                tx.To(),
	//	"contractAddress":   nil,
	//	"logs":              receipt.Logs,
	//	"logsBloom":         receipt.Bloom,
	//}
	//
	//// Assign receipt status or post state.
	//if len(receipt.PostState) > 0 {
	//	fields["root"] = hex.Bytes(receipt.PostState)
	//} else {
	//	fields["status"] = hex.Uint(receipt.Status)
	//}
	//if receipt.Logs == nil {
	//	fields["logs"] = [][]*transaction.Log{}
	//}
	//// If the ContractAddress is 20 0x0 bytes, assume it is not a contract creation
	//if receipt.ContractAddress != (types.Address{}) {
	//	fields["contractAddress"] = receipt.ContractAddress
	//}
	//return fields, nil

	return nil , nil
}

// sign is a helper function that signs a transaction with the private key of the given address.
func (s *PublicTransactionPoolAPI) sign(addr types.Address, tx *transaction.Transaction) (*transaction.Transaction, error) {
	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: addr}

	wallet, err := s.b.AccountManager().Find(account)
	if err != nil {
		return nil, err
	}
	// Request the wallet to sign the transaction
	chainID := s.b.ChainConfig().ChainId

	return wallet.SignTx(account, tx, chainID)
}

type SendTxAction struct {
	Address		*types.Address    `json:"address"`
	Params 		*hex.Bytes       `json:"params"`
}

// SendTxArgs represents the arguments to sumbit a new transaction into the transaction pool.
type SendTxArgs struct {
	From     types.Address  `json:"from"`
	Nonce    *hex.Uint64    `json:"nonce"`

	Actions  []SendTxAction    `json:"actions"`
}

// setDefaults is a helper function that fills in default values for unspecified tx fields.
func (args *SendTxArgs) setDefaults(ctx context.Context, b Backend) error {
	if args.Nonce == nil {
		nonce, err := b.GetPoolNonce(ctx, args.From)
		if err != nil {
			return err
		}
		args.Nonce = (*hex.Uint64)(&nonce)
	}
	if len(args.Actions) == 0 {
		return errors.New("no actions in transaction !!")
	}

	return nil
}

func (args *SendTxArgs) toTransaction() *transaction.Transaction {
	actions := []transaction.Action{}
	for _, argAction := range args.Actions {
		action := transaction.Action{argAction.Address, *argAction.Params}
		actions = append(actions, action)
	}
	return transaction.NewTransaction(uint64(*args.Nonce), actions)

	return nil
}

// submitTransaction is a helper function that submits tx to txPool and logs a message.
func submitTransaction(ctx context.Context, b Backend, tx *transaction.Transaction) (types.Hash, error) {
	if err := b.SendTx(ctx, tx); err != nil {
		return types.Hash{}, err
	}
	if len(tx.Data.Actions)==2 && tx.Data.Actions[1].Address == nil {
		signer := transaction.MakeSigner(b.ChainConfig(), b.CurrentBlock().Number())
		from, err := transaction.Sender(signer, tx)
		if err != nil {
			return types.Hash{}, err
		}
		addr := crypto.CreateAddress(from, tx.Nonce())
		logger.Info("Submitted contract creation", "fullhash", tx.Hash().Hex(), "contract", addr.Hex())
	} else {
		logger.Info("Submitted transaction", "fullhash", tx.Hash().Hex(), "recipient")
	}
	return tx.Hash(), nil
}

// SendTransaction creates a transaction for the given argument, sign it and submit it to the
// transaction pool.
func (s *PublicTransactionPoolAPI) SendTransaction(ctx context.Context, args SendTxArgs) (types.Hash, error) {

	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: args.From}

	wallet, err := s.b.AccountManager().Find(account)
	if err != nil {
		return types.Hash{}, err
	}

	if args.Nonce == nil {
		// Hold the addresse's mutex around signing to prevent concurrent assignment of
		// the same nonce to multiple accounts.
		s.nonceLock.LockAddr(args.From)
		defer s.nonceLock.UnlockAddr(args.From)
	}

	// Set some sanity defaults and terminate on failure
	if err := args.setDefaults(ctx, s.b); err != nil {
		return types.Hash{}, err
	}
	// Assemble the transaction and sign with the wallet
	tx := args.toTransaction()

	var chainID *big.Int
	chainID = s.b.ChainConfig().ChainId
	signed, err := wallet.SignTx(account, tx, chainID)
	if err != nil {
		return types.Hash{}, err
	}
	return submitTransaction(ctx, s.b, signed)
}

// SendRawTransaction will add the signed transaction to the transaction pool.
// The sender is responsible for signing the transaction and using the correct nonce.
func (s *PublicTransactionPoolAPI) SendRawTransaction(ctx context.Context, encodedTx hex.Bytes) (types.Hash, error) {
	tx := new(transaction.Transaction)
	byteBuf := bytes.NewBuffer(encodedTx)
	if err := msgp.Decode(byteBuf, tx); err != nil {
		return types.Hash{}, err
	}
	return submitTransaction(ctx, s.b, tx)
}

// Sign calculates an ECDSA signature for:
// keccack256("\x19MjoySigned Message:\n" + len(message) + message).
//
// Note, the produced signature conforms to the secp256k1 curve R, S and V values,
// where the V value will be 27 or 28 for legacy reasons.
//
// The account associated with addr must be unlocked.
//
func (s *PublicTransactionPoolAPI) Sign(addr types.Address, data hex.Bytes) (hex.Bytes, error) {
	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: addr}

	wallet, err := s.b.AccountManager().Find(account)
	if err != nil {
		return nil, err
	}
	// Sign the requested hash with the wallet
	signature, err := wallet.SignHash(account, signHash(data))
	if err == nil {
		signature[64] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper
	}
	return signature, err
}

// SignTransactionResult represents a Msgp encoded signed transaction.
type SignTransactionResult struct {
	Raw hex.Bytes                `json:"raw"`
	Tx  *transaction.Transaction `json:"tx"`
}

// SignTransaction will sign the given transaction with the from account.
// The node needs to have the private key of the account corresponding with
// the given from address and it needs to be unlocked.
func (s *PublicTransactionPoolAPI) SignTransaction(ctx context.Context, args SendTxArgs) (*SignTransactionResult, error) {
	if args.Nonce == nil {
		// Hold the addresse's mutex around signing to prevent concurrent assignment of
		// the same nonce to multiple accounts.
		s.nonceLock.LockAddr(args.From)
		defer s.nonceLock.UnlockAddr(args.From)
	}
	if err := args.setDefaults(ctx, s.b); err != nil {
		return nil, err
	}
	tx, err := s.sign(args.From, args.toTransaction())
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = msgp.Encode(&buf, tx)
	if err != nil {
		return nil, err
	}
	return &SignTransactionResult{buf.Bytes(), tx}, nil
}

// PendingTransactions returns the transactions that are in the transaction pool and have a from address that is one of
// the accounts this node manages.
func (s *PublicTransactionPoolAPI) PendingTransactions() ([]*RPCTransaction, error) {
	pending, err := s.b.GetPoolTransactions()
	if err != nil {
		return nil, err
	}

	transactions := make([]*RPCTransaction, 0, len(pending))
	for _, tx := range pending {
		var signer transaction.Signer = transaction.NewMSigner(tx.ChainId())

		from, _ := transaction.Sender(signer, tx)
		if _, err := s.b.AccountManager().Find(accounts.Account{Address: from}); err == nil {
			transactions = append(transactions, newRPCPendingTransaction(tx))
		}
	}
	return transactions, nil
}


func (s *PublicTransactionPoolAPI) Resend(ctx context.Context, sendArgs SendTxArgs) (types.Hash, error) {
	if sendArgs.Nonce == nil {
		return types.Hash{}, fmt.Errorf("missing transaction nonce in transaction spec")
	}
	if err := sendArgs.setDefaults(ctx, s.b); err != nil {
		return types.Hash{}, err
	}
	matchTx := sendArgs.toTransaction()
	pending, err := s.b.GetPoolTransactions()
	if err != nil {
		return types.Hash{}, err
	}

	for _, p := range pending {
		var signer transaction.Signer = transaction.NewMSigner(p.ChainId())

		wantSigHash := signer.Hash(matchTx)

		if pFrom, err := transaction.Sender(signer, p); err == nil && pFrom == sendArgs.From && signer.Hash(p) == wantSigHash {
			// Match. Re-sign and send the transaction.

			signedTx, err := s.sign(sendArgs.From, sendArgs.toTransaction())
			if err != nil {
				return types.Hash{}, err
			}
			if err = s.b.SendTx(ctx, signedTx); err != nil {
				return types.Hash{}, err
			}
			return signedTx.Hash(), nil
		}
	}

	return types.Hash{}, fmt.Errorf("Transaction %#x not found", matchTx.Hash())
}


