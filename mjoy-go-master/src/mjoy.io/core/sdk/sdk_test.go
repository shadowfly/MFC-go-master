package sdk

import (
	"testing"
	"mjoy.io/utils/database"
	"mjoy.io/common/types"
	"fmt"
)





func TestDataStore(t *testing.T){
	//create a database
	db,err := database.OpenMemDB()
	if err != nil {
		panic(err)
	}

	sdkHandler := NewTmpStatusManager(types.Hash{} , db)
	contractAddr := types.Address{}
	contractAddr[0] = 1

	accountAddr := types.Address{}
	Sys_SetValue(sdkHandler , contractAddr , accountAddr[:] , []byte{1,2,3,4,5})
	r := Sys_GetValue(sdkHandler , contractAddr , accountAddr[:] )
	fmt.Println("r:" , r)
}





