package intertypes

import (
	"mjoy.io/core/sdk"
	"mjoy.io/common/types"
	"mjoy.io/core/transaction"
)

//working result
type ActionResult struct {
	Key []byte
	Val []byte
}

type WorkResult struct {
	Err error
	Results []ActionResult
}
//GetResult
type GetResult struct {
	Err error
	Var []byte  //json result
}



type VmInterface interface {
	SendWork( types.Address ,  transaction.Action ,  *SystemParams)<-chan WorkResult
	GetStorage(address types.Address , action transaction.Action , params *SystemParams)GetResult
}

//SystemParams contain all system running params
type SystemParams struct {
	SdkHandler *sdk.TmpStatusManager    //contain current
	VmHandler VmInterface
}

func MakeSystemParams(sdkHandler *sdk.TmpStatusManager , vmHandler VmInterface )*SystemParams{
	s := new(SystemParams)
	s.SdkHandler = sdkHandler
	s.VmHandler = vmHandler
	return s
}


