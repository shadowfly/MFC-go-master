/*
All Innercontract implements should be added into InnerRegister slice.
*/

package interpreter

import (
	"mjoy.io/common/types"
	"mjoy.io/core/interpreter/balancetransfer"
)

type innerRegisterMap struct {
	address types.Address
	inner   InnerContract
}

type InnersRegister []innerRegisterMap

var allInnerRegister InnersRegister = InnersRegister{
	{balancetransfer.BalanceTransferAddress , balancetransfer.NewContractBalancer()},
}

