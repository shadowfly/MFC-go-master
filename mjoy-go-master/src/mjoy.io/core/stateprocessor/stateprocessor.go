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
// @File: stateprocessor.go
// @Date: 2018/05/08 15:18:08
////////////////////////////////////////////////////////////////////////////////

package stateprocessor

import (
	"mjoy.io/core/blockchain/block"
	"mjoy.io/core/state"
	"mjoy.io/params"
	"mjoy.io/core/transaction"
	"mjoy.io/common/types"
	"mjoy.io/utils/bloom"
	"mjoy.io/consensus"
	"mjoy.io/core/interpreter"
	"mjoy.io/utils/crypto"
	"mjoy.io/core/sdk"
	"mjoy.io/utils/database"
	"mjoy.io/core/interpreter/intertypes"
)

type IChainForState interface {
	consensus.ChainReader
}

// StateProcessor is a basic Processor, which takes care of transitioning
// state from one point to another.
//
// StateProcessor implements Processor.
type StateProcessor struct {
	config *params.ChainConfig // Chain configuration options
	cs     IChainForState      // chain interface for state processor
	engine consensus.Engine    // Consensus engine used for block rewards
}

type DbCache struct {
	Cache map[string]interpreter.MemDatabase
}

// NewStateProcessor initialises a new StateProcessor.
func NewStateProcessor(config *params.ChainConfig, cs IChainForState, engine consensus.Engine) *StateProcessor {
	return &StateProcessor{
		cs: cs,
		engine: engine,
		config: config,
	}
}

// Process processes the state changes according to the Mjoy rules by running
// the transaction messages using the statedb and applying any rewards to
// the processor (coinbase).
//
// Process returns the receipts and logs accumulated during the process.
// If any of the transactions failed  it will return an error.
func (p *StateProcessor) Process(blk *block.Block, statedb *state.StateDB, db database.IDatabaseGetter, config *params.ChainConfig) (*DbCache, transaction.Receipts, []*transaction.Log, error) {
	var (
		receipts transaction.Receipts
		header   = blk.Header()
		allLogs  []*transaction.Log
	)

	dbcache := &DbCache{
		Cache: make(map[string]interpreter.MemDatabase),
	}

	singner := block.NewBlockSigner(config.ChainId)
	coinbase, err := singner.Sender(header)
	if err != nil {
		logger.Error("Process: block signature is not right", err)
		return  nil, nil, nil, err
	}

	logger.Trace("Process: coinbase", coinbase.Hex())

	sdkHandler := sdk.NewTmpStatusManager(db, statedb , coinbase)
	vmHandler := interpreter.NewVm()
	//make sysparam

	sysparam := intertypes.MakeSystemParams(sdkHandler , vmHandler )


	// Iterate over and process the individual transactions
	for i, tx := range blk.Transactions() {
		statedb.Prepare(tx.Hash(), blk.Hash(), i)
		receipt, err := ApplyTransaction(p.config, nil, statedb, header, tx, dbcache , sysparam)
		if err != nil {
			logger.Errorf("ApplyTransacton Wrong.....:",err.Error())

			return  nil, nil, nil, err
		}
		receipts = append(receipts, receipt)
		allLogs = append(allLogs, receipt.Logs...)
	}


	// TODO: need to be compeleted, now skip this step
	// Finalize the block, applying any consensus engine specific extras (e.g. block rewards)
	if p.engine != nil {
		p.engine.Finalize(p.cs, header, statedb, blk.Transactions(), receipts, false)
	}

	return  dbcache, receipts, allLogs, nil
}


// ApplyTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. It returns the receipt
// for the transaction and an error if the transaction failed,
// indicating the block was invalid.
func ApplyTransaction(config *params.ChainConfig, author *types.Address, statedb *state.StateDB, header *block.Header, tx *transaction.Transaction, cache *DbCache , sysparam *intertypes.SystemParams) (*transaction.Receipt, error) {
	msg, err := tx.AsMessage(transaction.MakeSigner(config, &header.Number.IntVal))
	if err != nil {
		return nil, err
	}
	
	// Apply the transaction to the current state (included in the env)
	if author == nil {
		author = &header.BlockProducer
	}
	_, failed, err := ApplyMessage(statedb, msg, *author, cache, header,sysparam)
	if err != nil {
		return nil, err
	}
	// Update the state with pending changes
	statedb.Finalise(true)

	// Create a new receipt for the transaction, storing the intermediate root  by the tx
	// based on the mip phase, we're passing wether the root touch-delete accounts.
	receipt := transaction.NewReceipt(failed)
	receipt.TxHash = tx.Hash()
	// if the transaction created a contract, store the creation address in the receipt.
	if len(tx.Data.Actions) == 2 && tx.Data.Actions[1].Address == nil {
		receipt.ContractAddress = crypto.CreateAddress(msg.From(), tx.Nonce())
	}
	// Set the receipt logs and create a bloom for filtering
	receipt.Logs = statedb.GetLogs(tx.Hash())

	topics := []bloom.BloomByte{}
	for _, log := range receipt.Logs {
		topics = append(topics, log.Address)
		for _, topic := range log.Topics {
			topics = append(topics, topic)
		}
	}
	receipt.Bloom = bloom.CreateBloom(topics)

	return receipt, err
}
