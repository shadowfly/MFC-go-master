package consensus

import (
	"mjoy.io/core/blockchain/block"
	"mjoy.io/core/state"
	"mjoy.io/core/transaction"
	"math/big"
	"mjoy.io/common/types"
	"errors"
	"mjoy.io/common"
	"runtime"
	"crypto/ecdsa"
)


type Engine_basic struct {
	//todo: need interpreter information

	//key for sign header
	prv *ecdsa.PrivateKey
}

var (
	ErrBlockTime     = errors.New("timestamp less than or equal parent's timestamp")
	ErrSignature     = errors.New("signature is not right")
)

func NewBasicEngine(prv *ecdsa.PrivateKey)  (*Engine_basic){
	return &Engine_basic{
		prv,
	}
}

func (basic *Engine_basic)SetKey(prv *ecdsa.PrivateKey)  {
	basic.prv = prv
}

func (basic *Engine_basic) Author(chain ChainReader, header *block.Header) (types.Address, error) {
	singner := block.NewBlockSigner(chain.Config().ChainId)
	return singner.Sender(header)
}

func (basic *Engine_basic) VerifyHeader(chain ChainReader, header *block.Header, seal bool) error {
	//if the header is known, verify success
	number := header.Number.IntVal.Uint64()
	if chain.GetHeader(header.Hash(), number) != nil {
		return nil
	}

	parent := chain.GetHeader(header.ParentHash, number-1)
	if parent == nil {
		return ErrUnknownAncestor
	}

	// Verify that the block number is parent's +1
	if diff := new(big.Int).Sub(&header.Number.IntVal, &parent.Number.IntVal); diff.Cmp(common.Big1) != 0 {
		return ErrInvalidNumber
	}

	//verify time
	if header.Time.IntVal.Cmp(&parent.Time.IntVal) <= 0 {
		return ErrBlockTime
	}

	//verify ConsensusData
	if seal {
		if err := basic.VerifySeal(chain, header); err != nil {
			return err
		}
	}

	//verify signature
	singner := block.NewBlockSigner(chain.Config().ChainId)
	if _, err := singner.Sender(header); err!=nil{
		return ErrSignature
	}

	return nil
}

func (basic *Engine_basic) verifyHeader(chain ChainReader, header, parent *block.Header, seal bool) error {
	//if the header is known, verify success
	number := header.Number.IntVal.Uint64()
	if chain.GetHeader(header.Hash(), number) != nil {
		return nil
	}

	// Verify that the block number is parent's +1
	if diff := new(big.Int).Sub(&header.Number.IntVal, &parent.Number.IntVal); diff.Cmp(common.Big1) != 0 {
		return ErrInvalidNumber
	}

	//verify time
	if header.Time.IntVal.Cmp(&parent.Time.IntVal) <= 0 {
		return ErrBlockTime
	}

	//verify signature
	singner := block.NewBlockSigner(chain.Config().ChainId)
	if _, err := singner.Sender(header); err!=nil{
		return ErrSignature
	}

	return nil
}

func (basic *Engine_basic) verifyHeaderWorker(chain ChainReader, headers []*block.Header, seals []bool, index int) error {
	var parent *block.Header
	if index == 0 {
		parent = chain.GetHeader(headers[0].ParentHash, headers[0].Number.IntVal.Uint64()-1)
	} else if headers[index-1].Hash() == headers[index].ParentHash {
		parent = headers[index-1]
	}
	if parent == nil {
		return ErrUnknownAncestor
	}
	if chain.GetHeader(headers[index].Hash(), headers[index].Number.IntVal.Uint64()) != nil {
		return nil // known block
	}
	return basic.verifyHeader(chain, headers[index], parent,seals[index])
}

// VerifyHeaders is similar to VerifyHeader, but verifies a batch of headers
// concurrently. The method returns a quit channel to abort the operations and
// a results channel to retrieve the async verifications.
func (basic *Engine_basic) VerifyHeaders(chain ChainReader, headers []*block.Header, seals []bool) (chan<- struct{}, <-chan error) {
	workers := runtime.GOMAXPROCS(0)
	if len(headers) < workers {
		workers = len(headers)
	}
	// Create a task channel and spawn the verifiers
	var (
		inputs = make(chan int)
		done   = make(chan int, workers)
		errors = make([]error, len(headers))
		abort  = make(chan struct{})
	)
	for i := 0; i < workers; i++ {
		go func() {
			for index := range inputs {
				errors[index] = basic.verifyHeaderWorker(chain, headers, seals, index)
				done <- index
			}
		}()
	}

	errorsOut := make(chan error, len(headers))
	go func() {
		defer close(inputs)
		var (
			in, out = 0, 0
			checked = make([]bool, len(headers))
			inputs  = inputs
		)
		for {
			select {
			case inputs <- in:
				if in++; in == len(headers) {
					// Reached end of headers. Stop sending to workers.
					inputs = nil
				}
			case index := <-done:
				for checked[index] = true; checked[out]; out++ {
					errorsOut <- errors[out]
					if out == len(headers)-1 {
						return
					}
				}
			case <-abort:
				return
			}
		}
	}()
	return abort, errorsOut
}


//todo this need interpreter process ConsensusData
func (basic *Engine_basic) VerifySeal(chain ChainReader, header *block.Header) error {
	return nil
}

func (basic *Engine_basic) Prepare(chain ChainReader, header *block.Header) error {
	return nil
}


//todo this need interpreter process
//interpreter need change state
func (basic *Engine_basic) Finalize(chain ChainReader, header *block.Header, state *state.StateDB, txs []*transaction.Transaction, receipts []*transaction.Receipt, sign bool) (*block.Block, error) {
	//reward := big.NewInt(5e+18)
	//state.AddBalance(header.BlockProducer, reward)
	header.StateRootHash = state.IntermediateRoot()

	//sign header
	if sign {
		if basic.prv == nil {
			return nil, errors.New("No key found fo sign header")
		}
		blk := block.NewBlock(header, txs, receipts)

		err := block.SignHeaderInner(blk.B_header, block.NewBlockSigner(chain.Config().ChainId), basic.prv)
		if err != nil {
			return nil, err
		}
		return blk, nil
	} else {
		return block.NewBlock(header, txs, receipts), nil
	}

}

//todo fill header ConsensusData
func (basic *Engine_basic) Seal(chain ChainReader, block *block.Block, stop <-chan struct{}) (*block.Block, error){
	header := block.Header()
	return block.WithSeal(header), nil
}
