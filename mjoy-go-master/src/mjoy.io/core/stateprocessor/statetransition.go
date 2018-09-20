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
// @File: statetransition.go
// @Date: 2018/05/08 15:18:08
////////////////////////////////////////////////////////////////////////////////

package stateprocessor

import (
	"mjoy.io/common/types"
	"mjoy.io/core/state"
	"mjoy.io/core"
	"mjoy.io/core/transaction"
	"mjoy.io/core/interpreter"
	"mjoy.io/utils/crypto"
	"mjoy.io/core/blockchain/block"
	"mjoy.io/core/interpreter/intertypes"
)

/*
The State Transitioning Model

A state transition is a change made when a transaction is applied to the current world state
The state transitioning model does all all the necessary work to work out a valid new state root.

1) Nonce handling
3) Create a new state object if the recipient is \0*32
4) Value transfer
== If contract creation ==
  4a) Attempt to run transaction data
  4b) If valid, use result as code for the new state object
== end ==
5) Run Script section
6) Derive new state root
*/
type StateTransition struct {
	msg        Message
	actions     []transaction.Action
	statedb    *state.StateDB
	coinBase   types.Address
	Cache      *DbCache
	header      *block.Header
}

// Message represents a message sent to a contract.
type Message interface {
	From() types.Address
	Actions()[]transaction.Action
	Nonce() uint64
	CheckNonce() bool
}

// NewStateTransition initialises and returns a new state transition object.
func NewStateTransition(statedb *state.StateDB, msg Message, coinBase types.Address, cache *DbCache, header *block.Header ) *StateTransition {
	return &StateTransition{
		msg:      msg,
		actions:  msg.Actions(),
		statedb:  statedb,
		coinBase: coinBase,
		Cache : cache,
		header: header,
	}
}

// ApplyMessage computes the new state by applying the given message
// against the old state within the environment.
func ApplyMessage(statedb *state.StateDB, msg Message, coinBase types.Address, cache *DbCache,header *block.Header , sysparam *intertypes.SystemParams) ([]byte, bool, error) {
	return NewStateTransition(statedb, msg, coinBase, cache, header).TransitionDb(sysparam)
}

func (st *StateTransition) from() types.Address {
	f := st.msg.From()
	if !st.statedb.Exist(f) {
		st.statedb.CreateAccount(f)
	}
	return f
}

func (st *StateTransition) preCheck() error {
	msg := st.msg
	sender := st.from()

	// Make sure this transaction's nonce is correct
	if msg.CheckNonce() {
		nonce := st.statedb.GetNonce(sender)
		if nonce < msg.Nonce() {
			return core.ErrNonceTooHigh
		} else if nonce > msg.Nonce() {
			return core.ErrNonceTooLow
		}
	}
	return nil
}


// make log  function
func MakeLog(address types.Address, results interpreter.ActionResults, blockNumber uint64) *transaction.Log {
	topics := []types.Hash{}
	data := [][]byte{}

	for _, result := range results {
		topics = append(topics,types.BytesToHash(result.Key))
		data = append(data, result.Key)
	}

	return &transaction.Log{
		Address:     address,
		Topics:      topics,
		Data:        data,
		BlockNumber:   blockNumber,
	}
}

// TransitionDb will transition the state by applying the current message and
// returning the result. It returns an error if it
// failed. An error indicates a consensus issue.
func (st *StateTransition) TransitionDb(sysparam *intertypes.SystemParams) (ret []byte, failed bool, err error) {
	if err = st.preCheck(); err != nil {
		return
	}

	sender := st.from() // err checked in preCheck

	contractCreation := false

	if len(st.actions) == 2 && st.actions[1].Address == nil{
		contractCreation = true
	}
	// Snapshot !!!!!!!!!!!!!!!!!
	snapshot := st.statedb.Snapshot()

	resultMem := []*interpreter.MemDatabase{}

	if contractCreation {
		results, _, err := interpreter.Create(sender, st.statedb, st.actions)
		if err != nil {
			st.statedb.RevertToSnapshot(snapshot)
			return nil, true, err
		}
		for _, res := range results {
			resM := &interpreter.MemDatabase{*st.actions[0].Address, res.Key, res.Val}
			resultMem = append(resultMem, resM)
		}
		// make log for receipt
		log := MakeLog(*st.actions[0].Address, results, st.header.Number.IntVal.Uint64())
		st.statedb.AddLog(log)
	} else {
		logger.Debugf("Just process actions transaction.")
		st.statedb.SetNonce(sender, st.statedb.GetNonce(sender) +1 )
		for _,action := range st.actions {
			//resulst := make(chan interpreter.WorkResult)
			resulstChan :=sysparam.VmHandler.SendWork(sender,action,sysparam)

			result := <-resulstChan
			if result.Err != nil {
				logger.Error("action fail.", result.Err)
				st.statedb.RevertToSnapshot(snapshot)
				return nil, true, err
			}
			for _, res := range result.Results {
				resM := &interpreter.MemDatabase{*action.Address, res.Key, res.Val}
				resultMem = append(resultMem, resM)
			}
			// make log for receipt
			log := MakeLog(*action.Address, result.Results, st.header.Number.IntVal.Uint64())
			st.statedb.AddLog(log)
		}
	}

	for _, result := range resultMem {
		storgageKey := append(result.Address.Bytes(), result.Key...)

		//1, change statedb storage
		storageKeyHash := crypto.Keccak256Hash(storgageKey)
		storageValHash := crypto.Keccak256Hash(result.Val)
		st.statedb.SetState(result.Address, storageKeyHash, storageValHash)

		//2, collect results for block producer future write level db
		st.Cache.Cache[string(storgageKey)] = interpreter.MemDatabase{
			result.Address,
			storageValHash.Bytes(),
			result.Val}
	}

	return ret, false, err
}
