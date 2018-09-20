package interpreter

import (
	"errors"
	"sync"
	"mjoy.io/core/transaction"
	"mjoy.io/common/types"
	"time"
	"fmt"
	"mjoy.io/core/interpreter/intertypes"
	"mjoy.io/core/interpreter/balancetransfer"
)

//Test addressd
var BalanceTransferAddress  = types.HexToAddress("0x0000000000000000000000000000000000000001")

func NewVm()*Vms{
	vm := new(Vms)
	vm.init()
	return vm
}

type Vms struct {
	pInnerContractMaper *InnerContractManager

	lock sync.RWMutex   //working  mux
	WorkingChan chan *Work    //work chan
	exit        chan struct{}

}


func (this *Vms)init(){
	//init workingChan
	this.WorkingChan = make(chan *Work , 1000)
	//exit chan
	this.exit = make(chan struct{} , 1)
	//init innerContractMapper
	this.pInnerContractMaper = NewInnerContractManager()
	//go this.Run()

}


/********************************************************************/
//Deal Actions..........
/********************************************************************/
//DealActons is a full work,and return a workresult to caller,if get one err ,return
func (this *Vms)DealActions(pWork *Work)error{
	var workResult intertypes.WorkResult
	fmt.Println("DealActions........")
	workResult.Err = nil
	for _ , a := range pWork.actions{
		workResult.Results = make([]intertypes.ActionResult , 0 )
		workResult.Err = nil

		r , err := this.DealAction(pWork.contractAddress,a ,pWork.sysParams)
		if err != nil{
			//get a err,return
			workResult.Results = nil
			workResult.Err = err
			pWork.resultChan <- workResult
			return err
		}

		//get a result
		workResult.Results = append(workResult.Results , r...)
		pWork.resultChan <- workResult
	}

	return nil
}

//DealActions is a little part of full work
func (this *Vms)DealAction(contractAddress types.Address , action transaction.Action ,sysparam *intertypes.SystemParams)([]intertypes.ActionResult , error){
	if this.pInnerContractMaper.Exist(contractAddress){
		results , err := this.pInnerContractMaper.DoFun(contractAddress , action.Params , sysparam)
		if err != nil {
			return nil , err
		}
		return results , nil
	}
	return nil , errors.New("innerContract Not Exist....")
}

/********************************************************************/
//Deal Work..........
/********************************************************************/
//SendWork is called when applytransaction
func (this *Vms)SendWork(from types.Address , action transaction.Action , sysParam *intertypes.SystemParams)<-chan intertypes.WorkResult{

	actions := []transaction.Action{}
	actions = append(actions , action)
	fmt.Println("SendWork actions Len:" , len(actions))
	w := NewWork(*action.Address , actions , sysParam)
	//this.WorkingChan<-w

	go this.DealActions(w)

	fmt.Println("SendWork.....")

	return w.resultChan
}

func (this *Vms)GetStorage(address types.Address , action transaction.Action , sysParam *intertypes.SystemParams)intertypes.GetResult{
	actions := []transaction.Action{}
	actions = append(actions , action)
	fmt.Println("SendWork actions Len:" , len(actions))
	w := NewWork(*action.Address , actions , sysParam)
	//this.WorkingChan<-w

	go this.DealActions(w)
	wRslt := <-w.resultChan
	getRslt := intertypes.GetResult{}
	getRslt.Err = wRslt.Err
	if getRslt.Err != nil {
		getRslt.Var = nil
	}else{
		getRslt.Var = wRslt.Results[0].Val
	}
	return getRslt
}


/********************************************************************/
//cycle Dealing
/********************************************************************/
func (this *Vms)Run(){
	//go this.TestRun()

	for{
		select  {
		case newWork := <-this.WorkingChan:
			go this.DealActions(newWork)
		case <-this.exit:
			return

		}
	}
}


func (this *Vms)TestRun(){
	for{
		time.Sleep(4*time.Second)
		fmt.Println("For Vm testing print.......")
	}
}


//will uesed by txpool
func GetPriority(from types.Address , actions []transaction.Action)int{
	//some calculation for priority
	priority , err :=balancetransfer.CheckFee(*actions[0].Address , actions[0].Params)
	if err != nil {
		logger.Errorf("Get Err When call GetPriority:" , err.Error())
	}
	return priority
}

















