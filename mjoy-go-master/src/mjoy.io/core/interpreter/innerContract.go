/*
This file define a InnerContract interface
*/

package interpreter

import (
	"sync"
	"mjoy.io/common/types"
	"math/big"
	"mjoy.io/core/interpreter/intertypes"
	"fmt"
)

//InnerContrancInterface
type InnerContract interface {
	DoFun( params []byte , sysparam *intertypes.SystemParams)([]intertypes.ActionResult , error)

}

//InnerContranctMap is a innerContract controller,like check contract ,do a contract
type InnerContractManager struct {
	mu sync.RWMutex
	Inners map[types.Address]InnerContract
}

//New A InnerContractMaper
func NewInnerContractManager()*InnerContractManager{
	maper := new(InnerContractManager)
	maper.Inners = make(map[types.Address]InnerContract)
	maper.init()
	return maper
}

func (this *InnerContractManager)init(){
	this.register()
}

var zeroAddress = types.BigToAddress(big.NewInt(0))

func (this *InnerContractManager)register(){
	this.mu.Lock()
	defer this.mu.Unlock()

	if len(allInnerRegister) == 0 {
		logger.Error("InnerContractMaper register len == 0")
		return
	}

	for _ , obj := range allInnerRegister {
		fmt.Println("registerAddr :" , obj.address.Hex())
		this.Inners[obj.address] = obj.inner
		//if obj.address != zeroAddress{
		//	this.Inners[obj.address] = obj.inner
		//}
	}
}


//check innerContract is exist
func (this *InnerContractManager)Exist(address types.Address)bool{
	this.mu.RLock()
	defer this.mu.RUnlock()

	if _ , ok := this.Inners[address];ok{
		return true
	}
	return false
}

//call a innerContract.Please call Exist ensure a innerContract is exist or not before this
func (this *InnerContractManager)DoFun(address types.Address , params []byte,sysparam *intertypes.SystemParams)([]intertypes.ActionResult , error){
	inner := this.Inners[address]
	return inner.DoFun(params,sysparam)
}




