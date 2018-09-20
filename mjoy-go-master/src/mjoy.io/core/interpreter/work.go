package interpreter

import (
	"mjoy.io/core/transaction"
	"mjoy.io/common/types"
	"mjoy.io/core/interpreter/intertypes"
)


//a operate work

type Work struct {
	actions []transaction.Action
	contractAddress types.Address
	sysParams *intertypes.SystemParams
	resultChan chan intertypes.WorkResult
}

func NewWork(contractAddress types.Address , actions []transaction.Action  , sysParams *intertypes.SystemParams)*Work{
	w := new(Work)

	//copy actions
	w.contractAddress = contractAddress     //who deal the transaction
	w.actions= make([]transaction.Action , len(actions))
	w.sysParams = sysParams
	copy(w.actions , actions)
	w.resultChan = make(chan intertypes.WorkResult , 1)
	return w
}







