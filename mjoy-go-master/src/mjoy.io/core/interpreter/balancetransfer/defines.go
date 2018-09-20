package balancetransfer

import (
	"encoding/json"
	"mjoy.io/common/types"
	"fmt"
)

//here for test,do not add msgp
type BalanceValue struct {
	Amount int    `json:"amount"`
}



//Balance Check Result
type AccountBalance struct {
	Address string   `json:"address"`
	Amount int              `json:"amount"`
}

type AccountsBalance struct {
	All []AccountBalance    `json:"all"`
}


func MakeActionParamsReword(producer types.Address)[]byte{
	a := make(map[string]interface{})
	a["funcId"] = "1"

	a["producer"] = producer.Hex()

	r , err :=json.Marshal(a)
	if err != nil {
		return nil
	}
	return r
}


//for inner test
func MakaBalanceTransferParam(from , to types.Address , amount int)[]byte{
	a := make(map[string]interface{})

	a["funcId"] = "0"
	a["from"] = from.Hex()
	a["to"] = to.Hex()
	a["amount"] = fmt.Sprintf("%d" , amount)

	r , err :=json.Marshal(a)
	if err != nil {
		return nil
	}
	return r
}

