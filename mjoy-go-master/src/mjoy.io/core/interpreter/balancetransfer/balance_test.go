package balancetransfer

import (
	"testing"
	"encoding/json"
	"fmt"
	"mjoy.io/core/interpreter/intertypes"
	"errors"
)

type Para struct {
	FuncName string
	Para map[string]interface{}
}

func TestJosnInterface(t *testing.T) {
	para := &Para{}
	para.Para = make(map[string]interface{})

	para.Para["from"] = "0x214343434343"
	para.Para["to"] = "0x2222222222"
	para.Para["amount"] = "100"
	para.FuncName = "TransferBalance"

	jsonResult,err := json.Marshal(para)
	fmt.Println(jsonResult,err)


	paraUnmarshal := &Para{}

	err = json.Unmarshal(jsonResult , paraUnmarshal)
	fmt.Println(paraUnmarshal,err)

}

type TestRequest struct {
	FuncId  string `json:"funcId"`
	Addresses []string  `json:"addresses"`

}


func GetBalanceForTest(param map[string]interface{} )([]intertypes.ActionResult , error){


	logger.Info(">>>>>>>>GetBalance.............................................")
	allRequestAddress := []string{}
	//get params

	if addressesInterface , ok := param["addresses"];!ok{
		//errDeal
		return nil , errors.New("GetBalance:No requests....")
	}else{
		sliceInterface,ok := addressesInterface.([]interface{})
		if !ok {
			return nil,errors.New("requests's type not map[string]interface{}")
		}

		for _ , v := range sliceInterface {
			addr , ok := v.(string)
			if !ok{
				return nil , errors.New("sliceInterface element is not string ")
			}
			allRequestAddress = append(allRequestAddress , addr)

		}
	}

	balanceCheckResult := new(AccountsBalance)

	for _ , v := range allRequestAddress {
		fmt.Printf("WeParsed Address:%s\n" , v)
		//check Balance
		balanceCheckVal := new(BalanceValue)
		balanceCheckVal.Amount = 10
		balanceCheckResult.All = append(balanceCheckResult.All , AccountBalance{v ,balanceCheckVal.Amount })


	}
	//make a result
	results := make([]intertypes.ActionResult ,0, 1)
	//Marshal balanceCheckResult to []byte
	resultBytes  , err := json.Marshal(balanceCheckResult)
	//fill result
	if err != nil || resultBytes == nil {

		return nil , errors.New(fmt.Sprintf("GetBalance Last Marshal Err:%s" , err.Error()))
	}
	//right
	results = append(results , intertypes.ActionResult{nil , resultBytes})
	return results , nil

}


func TestBalanceCheckJsonAndParse(t *testing.T){

	k := new(TestRequest)
	k.FuncId = "1"
	k.Addresses = append(k.Addresses , "0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed")
	k.Addresses = append(k.Addresses , "0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed")
	k.Addresses = append(k.Addresses , "0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed")
	k.Addresses = append(k.Addresses , "0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed")
	k.Addresses = append(k.Addresses , "0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed")

	marshalResults , err := json.Marshal(k)
	if err != nil {
		panic(err)
	}
	fmt.Printf("marshalResults:%s\n" ,string(marshalResults))

	//unmarshal to a map
	m := make(map[string]interface{})

	err = json.Unmarshal(marshalResults , &m)
	if err != nil {
		panic(err)
	}

	allResult , err := GetBalanceForTest(m )
	fmt.Println(allResult)


}