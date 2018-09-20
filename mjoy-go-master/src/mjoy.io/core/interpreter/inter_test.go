package interpreter

import (
	"mjoy.io/core/interpreter/balancetransfer"
	"encoding/json"
	"mjoy.io/utils/database"
	"mjoy.io/core/sdk"
	"mjoy.io/common/types"
	"fmt"
	"reflect"
	"testing"
	"mjoy.io/core/transaction"
	"mjoy.io/core/interpreter/intertypes"
)

func checkResultsData(sdkHandler *sdk.TmpStatusManager){
	contractAddr := types.Address{}
	contractAddr[0] = 1

	fromAddr := types.Address{}
	fromAddr[2] = 1

	refind := sdk.Sys_GetValue(sdkHandler , balancetransfer.BalanceTransferAddress , fromAddr[:])
	if refind == nil{
		fmt.Println("not find data just store before")
		return
	}
	a := new(balancetransfer.BalanceValue)
	err := json.Unmarshal(refind , a)
	if err != nil {
		fmt.Println("err:" , err)
	}

	fmt.Println("account 1 balance:" , a.Amount)

	toAddr := types.Address{}
	toAddr[3] = 1
	refind = sdk.Sys_GetValue(sdkHandler , balancetransfer.BalanceTransferAddress , toAddr[:])
	if refind == nil{
		fmt.Println("not find data just store before")
		return
	}

	err = json.Unmarshal(refind , a)
	if err != nil {
		fmt.Println("err:" , err)
	}

	fmt.Println("account 2 balance:" , a.Amount)
}


func makeTestData()*sdk.TmpStatusManager{

	//init database
	db,err := database.OpenMemDB()
	if err != nil {
		panic(err)
	}

	a := new(balancetransfer.BalanceValue)
	a.Amount = 1000

	lastAccountInfoData , err := json.Marshal(a)
	if err != nil {
		fmt.Println("json marshal wrong..........err:",err)
		return nil
	}
	//store the data
	sdkHandler := sdk.NewTmpStatusManager(types.Hash{} , db)
	contractAddr := types.Address{}
	contractAddr[0] = 1

	accountAddr := types.Address{}
	accountAddr[2] = 1
	sdk.Sys_SetValue(sdkHandler , balancetransfer.BalanceTransferAddress , accountAddr[:] , lastAccountInfoData)
	//sdk.PtrSdkManager.Down()
	refind := sdk.Sys_GetValue(sdkHandler , balancetransfer.BalanceTransferAddress , accountAddr[:])
	if refind == nil{
		fmt.Println("not find data just store before")
	}
	fmt.Printf("get store data before :%x\n" , refind)
	return sdkHandler
}

func makeActionParams()[]byte{
	a := make(map[string]interface{})
	a["funcId"] = "0"

	fromAddr := types.Address{}
	fromAddr[2] = 1
	a["from"] = fromAddr.Hex()

	toAddr := types.Address{}
	toAddr[3] = 1
	a["to"] = toAddr.Hex()

	a["amount"] = int64(10)

	fmt.Println("type amount:" , reflect.TypeOf(a["amount"]))
	r , err :=json.Marshal(a)
	if err != nil {
		return nil
	}
	return r
}

func makeActionParamsReword()[]byte{
	a := make(map[string]interface{})
	a["funcId"] = "1"

	fromAddr := types.Address{}
	fromAddr[2] = 1
	a["producer"] = fromAddr.Hex()

	r , err :=json.Marshal(a)
	if err != nil {
		return nil
	}
	return r
}

/*
todo: Focus these steps below
step 1:hold the tmp status manager
	sdkHandler := sdk.NewTmpStatusManager(types.Hash{} , db)

step 2:create a Vm
	pNewVm := NewVm()

step 3:create sysparam
	sysparam := intertypes.MakeSystemParams(sdkHandler)

step 4:send the action to Vm with sysparam(sdkHandler,the vm must read data from the sdkHandler)
	pNewVm.SendWork(types.Address{} , action , sysparam)

step 5:deal the results
*/

func TestInterDbNoDataBefore(t *testing.T){
	//make test data
	sdkHandler := makeTestData()
	checkResultsData(sdkHandler)
	action:= transaction.Action{}
	contranctAddr := balancetransfer.BalanceTransferAddress
	action.Address = &contranctAddr
	action.Params = makeActionParams()

	pNewVm := NewVm()
	//sdk.PtrSdkManager.Prepare(types.Hash{})
	fmt.Println("Start Testing....")


	//create system params
	sysparam := intertypes.MakeSystemParams(sdkHandler , pNewVm)
	rChan := pNewVm.SendWork(types.Address{} , action , sysparam)
	rw := <-rChan
	fmt.Println("get A result")
	fmt.Println("resultsLen :" , len(rw.Results))
	fmt.Println("err:" , rw.Err)
	_ = rw
	checkResultsData(sdkHandler)

	action.Params = makeActionParamsReword()
	rChan = pNewVm.SendWork(types.Address{} , action , sysparam)
	rw = <-rChan
	//time.Sleep(1*time.Second)
	checkResultsData(sdkHandler)

}