package balancetransfer

import (
	"mjoy.io/common/types"
	"strconv"
	"errors"
)

func CheckFee(addr types.Address , params []byte)(int , error){
	if addr != BalanceTransferAddress {
		return 0 , errors.New("Contract address wrong")
	}
	jsonParams , err := ParseParms(params)
	if err != nil {
		return 0 , err
	}
	if v,ok := jsonParams["funcId"];!ok {
		return 0 , errors.New("NoFuncId for CheckFee")
	}else{
		vi,err := strconv.Atoi(v.(string))
		if err != nil {
			return 0 , err
		}
		if vi != TransferFee_FunId {
			return 0 , errors.New("funcId != CheckFee funcId")
		}
		//parse amounts
		if amountStr , ok := jsonParams["amount"];!ok {
			return 0 , errors.New("No Amount in CheckFee params")
		}else{
			amountFee,err := strconv.Atoi(amountStr.(string))
			if err != nil {
				return 0 , err
			}
			return amountFee , nil
		}
	}

	return 0 , errors.New("CheckFee Unkown Error")

}
