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
// @File: blockproducer.go
// @Date: 2018/05/08 17:22:08
////////////////////////////////////////////////////////////////////////////////

// Package blockproducer implements Mjoy block creation and producing block.
package blockproducer

import (
	"fmt"
	"sync/atomic"

	"mjoy.io/accounts"
	"mjoy.io/common/types"
	"mjoy.io/consensus"

	"mjoy.io/core/state"
	"mjoy.io/node/services/mjoy/downloader"

	"mjoy.io/utils/event"
	"mjoy.io/params"

	"mjoy.io/core/blockchain/block"
	"mjoy.io/utils/database"
	"mjoy.io/core/blockchain"
	"mjoy.io/core/txprocessor"
	"mjoy.io/core/transaction"
	"mjoy.io/core/interpreter/intertypes"
)


// Backend wraps all methods required for producing block.
type Backend interface {
	AccountManager() *accounts.Manager
	BlockChain() *blockchain.BlockChain
	TxPool() *txprocessor.TxPool
	ChainDb() database.IDatabase
}
//Interpreter
type Interpreter interface {
	SendWork(from types.Address , action transaction.Action , systemparam *intertypes.SystemParams)<-chan intertypes.WorkResult
}
// Blockproducer creates blocks and searches for proof-of-work values.
type Blockproducer struct {
	mux *event.TypeMux

	producer *producer

	coinbase types.Address
	producing   int32
	mjoy      Backend
	inter     Interpreter
	engine   consensus.Engine

	canStart    int32 // can start indicates whether we can start the producing operation
	shouldStart int32 // should start indicates whether we should start after sync
}

func New(mjoy Backend,inter Interpreter, config *params.ChainConfig, mux *event.TypeMux, engine consensus.Engine) *Blockproducer {
	blockproducer := &Blockproducer{
		mjoy:      mjoy,
		inter:      inter,
		mux:      mux,
		engine:   engine,
		producer:   newProducer(config, engine, types.Address{}, mjoy,inter, mux),
		canStart: 1,
	}
	blockproducer.Register(NewCpuAgent(mjoy.BlockChain(), engine))
	go blockproducer.update()

	return blockproducer
}

// update keeps track of the downloader events. Please be aware that this is a one shot type of update loop.
// It's entered once and as soon as `Done` or `Failed` has been broadcasted the events are unregistered and
// the loop is exited. This to prevent a major security vuln where external parties can DOS you with blocks
// and halt your producing operation for as long as the DOS continues.


func (self *Blockproducer) update() {
	events := self.mux.Subscribe(downloader.StartEvent{}, downloader.DoneEvent{}, downloader.FailedEvent{})
out:
	for ev := range events.Chan() {
		switch ev.Data.(type) {
		case downloader.StartEvent:
			atomic.StoreInt32(&self.canStart, 0)
			if self.Producing() {
				self.Stop()
				atomic.StoreInt32(&self.shouldStart, 1)
				logger.Info("Producing aborted due to sync")
			}
		case downloader.DoneEvent, downloader.FailedEvent:
			shouldStart := atomic.LoadInt32(&self.shouldStart) == 1

			atomic.StoreInt32(&self.canStart, 1)
			atomic.StoreInt32(&self.shouldStart, 0)
			if shouldStart {
				self.Start(self.coinbase)
			}
			// unsubscribe. we're only interested in this event once
			events.Unsubscribe()
			// stop immediately and ignore all further pending events

			break out
		}
	}
}

func (self *Blockproducer) Start(coinbase types.Address) {
	atomic.StoreInt32(&self.shouldStart, 1)

	self.producer.setCoinbase(coinbase)
	self.coinbase = coinbase

	if atomic.LoadInt32(&self.canStart) == 0 {
		logger.Info("Network syncing, will start blockproducer afterwards")
		return
	}
	atomic.StoreInt32(&self.producing, 1)

	logger.Info("Starting producing operation")
	self.producer.start()
	self.producer.commitNewWork()
}

func (self *Blockproducer) Stop() {
	self.producer.stop()
	atomic.StoreInt32(&self.producing, 0)
	atomic.StoreInt32(&self.shouldStart, 0)
}

func (self *Blockproducer) Register(agent Agent) {
	if self.Producing() {
		agent.Start()
	}
	self.producer.register(agent)
}

func (self *Blockproducer) Unregister(agent Agent) {
	self.producer.unregister(agent)
}

func (self *Blockproducer) Producing() bool {
	return atomic.LoadInt32(&self.producing) > 0
}

func (self *Blockproducer) HashRate() (tot int64) {
	//if pow, ok := self.engine.(consensus.PoW); ok {
	//	tot += int64(pow.Hashrate())
	//}
	// do we care this might race? is it worth we're rewriting some
	// aspects of the worker/locking up agents so we can get an accurate
	// hashrate?
	for agent := range self.producer.agents {
		if _, ok := agent.(*CpuAgent); !ok {
			tot += agent.GetHashRate()
		}
	}
	return
}

func (self *Blockproducer) SetExtra(extra []byte) error {
	if uint64(len(extra)) > params.MaximumExtraDataSize {
		return fmt.Errorf("Extra exceeds max length. %d > %v", len(extra), params.MaximumExtraDataSize)
	}
	self.producer.setExtra(extra)
	return nil
}

// Pending returns the currently pending block and associated state.
func (self *Blockproducer) Pending() (*block.Block, *state.StateDB) {
	return self.producer.pending()
}

// PendingBlock returns the currently pending block.
//
// Note, to access both the pending block and the pending state
// simultaneously, please use Pending(), as the pending state can
// change between multiple method calls
func (self *Blockproducer) PendingBlock() *block.Block {
	return self.producer.pendingBlock()
}

func (self *Blockproducer) SetCoinbase(addr types.Address) {
	self.coinbase = addr
	self.producer.setCoinbase(addr)
}
