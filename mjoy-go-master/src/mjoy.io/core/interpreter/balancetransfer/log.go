package balancetransfer

import (
	"mjoy.io/log"
	"fmt"
	"os"
)

var (
	logTag = "interpreter.balancetransfer"
	logger log.Logger
)



func init() {
	logger = log.GetLogger(logTag)
	if logger == nil {
		fmt.Errorf("Can not get logger(%s)\n", logTag)
		os.Exit(1)
	}
}