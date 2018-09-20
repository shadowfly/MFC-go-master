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
// @File: api_backend.go
// @Date: 2018/05/08 18:02:08
////////////////////////////////////////////////////////////////////////////////

package mjoy

import (
	"context"

	"mjoy.io/params"
	"mjoy.io/core/state"
	"mjoy.io/core"
	"mjoy.io/utils/event"
	"mjoy.io/node/services/mjoy/downloader"
	"mjoy.io/accounts"
	"mjoy.io/core/blockchain/block"
	"mjoy.io/communication/rpc"
	"mjoy.io/common/types"
	"mjoy.io/core/transaction"
	"mjoy.io/core/blockchain"
	"mjoy.io/utils/database"
	"mjoy.io/utils/bloom"
)

// MjoyApiBackend implements mjoyapi.Backend for full nodes
type MjoyApiBackend struct {
	mjoy *Mjoy
}

func (b *MjoyApiBackend) ChainConfig() *params.ChainConfig {
	return b.mjoy.chainConfig
}

func (b *MjoyApiBackend) CurrentBlock() *block.Block {
	return b.mjoy.blockchain.CurrentBlock()
}

func (b *MjoyApiBackend) SetHead(number uint64) {
	b.mjoy.protocolManager.downloader.Cancel()
	b.mjoy.blockchain.SetHead(number)
}

func (b *MjoyApiBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*block.Header, error) {

	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.mjoy.blockchain.CurrentBlock().Header(), nil
	}
	return b.mjoy.blockchain.GetHeaderByNumber(uint64(blockNr)), nil
}

func (b *MjoyApiBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*block.Block, error) {

	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.mjoy.blockchain.CurrentBlock(), nil
	}
	return b.mjoy.blockchain.GetBlockByNumber(uint64(blockNr)), nil
}

func (b *MjoyApiBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *block.Header, error) {

	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := b.mjoy.BlockChain().StateAt(header.StateRootHash)
	return stateDb, header, err
}

func (b *MjoyApiBackend) GetBlock(ctx context.Context, blockHash types.Hash) (*block.Block, error) {
	return b.mjoy.blockchain.GetBlockByHash(blockHash), nil
}

func (b *MjoyApiBackend) GetReceipts(ctx context.Context, blockHash types.Hash) (transaction.Receipts, error) {
	return blockchain.GetBlockReceipts(b.mjoy.chainDb, blockHash, blockchain.GetBlockNumber(b.mjoy.chainDb, blockHash)), nil
}

func (b *MjoyApiBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.mjoy.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *MjoyApiBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.mjoy.BlockChain().SubscribeChainEvent(ch)
}

func (b *MjoyApiBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.mjoy.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *MjoyApiBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return b.mjoy.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *MjoyApiBackend) SubscribeLogsEvent(ch chan<- []*transaction.Log) event.Subscription {
	return b.mjoy.BlockChain().SubscribeLogsEvent(ch)
}

func (b *MjoyApiBackend) SendTx(ctx context.Context, signedTx *transaction.Transaction) error {
	return b.mjoy.txPool.AddLocal(signedTx)
}

func (b *MjoyApiBackend) GetPoolTransactions() (transaction.Transactions, error) {
	pending, err := b.mjoy.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs transaction.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *MjoyApiBackend) GetPoolTransaction(hash types.Hash) *transaction.Transaction {
	return b.mjoy.txPool.Get(hash)
}

func (b *MjoyApiBackend) GetPoolNonce(ctx context.Context, addr types.Address) (uint64, error) {
	return b.mjoy.txPool.State().GetNonce(addr), nil
}

func (b *MjoyApiBackend) Stats() (pending int, queued int) {
	return b.mjoy.txPool.Stats()
}

func (b *MjoyApiBackend) TxPoolContent() (map[types.Address]transaction.Transactions, map[types.Address]transaction.Transactions) {
	return b.mjoy.TxPool().Content()
}

func (b *MjoyApiBackend) SubscribeTxPreEvent(ch chan<- core.TxPreEvent) event.Subscription {
	return b.mjoy.TxPool().SubscribeTxPreEvent(ch)
}

func (b *MjoyApiBackend) Downloader() *downloader.Downloader {
	return b.mjoy.Downloader()
}

func (b *MjoyApiBackend) ProtocolVersion() int {
	return b.mjoy.MjoyVersion()
}

func (b *MjoyApiBackend) ChainDb() database.IDatabase {
	return b.mjoy.ChainDb()
}

func (b *MjoyApiBackend) EventMux() *event.TypeMux {
	return b.mjoy.EventMux()
}

func (b *MjoyApiBackend) AccountManager() *accounts.Manager {
	return b.mjoy.AccountManager()
}

func (b *MjoyApiBackend) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.mjoy.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

func (b *MjoyApiBackend) ServiceFilter(ctx context.Context, session *bloom.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.mjoy.bloomRequests)
	}
}
