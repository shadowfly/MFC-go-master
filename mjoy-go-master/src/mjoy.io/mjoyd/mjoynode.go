////////////////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 The mjoy-go Authors.
//
// The mjoy-go is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
// @File: mjoynode.go
// @Date: 2018/05/08 15:18:08
////////////////////////////////////////////////////////////////////////////////

package main

import (
	"mjoy.io/utils/crypto"
	"mjoy.io/node"
	"mjoy.io/communication/p2p"
	"mjoy.io/node/services/mjoy"
	"mjoy.io/core/genesis"
	"mjoy.io/common/types"
	"mjoy.io/core/txprocessor"
	"fmt"
	"os"
	"os/signal"
	"gopkg.in/urfave/cli.v1"
	"mjoy.io/mjoyd/utils"
	//"mjoy.io/mjoyd/defaults"
	"strings"
	"mjoy.io/communication/p2p/discover"
	"mjoy.io/params"
	"mjoy.io/mjoyd/config"
)

var (
	testNodeKey , _ = crypto.GenerateKey()
)

func testNodeConfig()*node.Config{
	return &node.Config{
		Name:"testNode",
		P2P:p2p.Config{PrivateKey:testNodeKey},
	}
}

var Cfg mjoy.MjoydConfig = mjoy.MjoydConfig{
	Mjoy:mjoy.Config{
		Genesis:genesis.DefaultGenesisBlock(),
		NetworkId:100,
		Coinbase:types.Address{},
		TxPool:txprocessor.DefaultTxPoolConfig,
	},
	Node:*testNodeConfig(),
}

func createMjoydCfg(ctx *cli.Context)*mjoy.MjoydConfig{
	//modules
	nodeModulesStr := ctx.GlobalString(utils.HttpModulesFlag.Name)
	nodeModules := strings.Split(nodeModulesStr , ",")
	//http host

	//http port
	nodeConfig := node.Config{
		Name:fmt.Sprintf("%s_node", ctx.GlobalString(utils.NodeNameFlag.Name)),
		DataDir:ctx.GlobalString(utils.DataDirFlag.Name),
		KeyStoreDir:ctx.GlobalString(utils.KeysotreFlag.Name),
		HTTPHost:ctx.GlobalString(utils.HttpHostFlag.Name),
		HTTPPort:ctx.GlobalInt(utils.HttpPortFlag.Name),
		P2P:p2p.Config{
			MaxPeers:10,
			Name:ctx.GlobalString(utils.NodeNameFlag.Name),
			ListenAddr:fmt.Sprintf(":%d",ctx.GlobalInt(utils.ListenPortFlag.Name)),
			DiscoveryV5:true,
			DiscoveryV5Addr:fmt.Sprintf(":%d",ctx.GlobalInt(utils.ListenPortFlag.Name)+1),
		},

	}
	//http Modules set
	if len(nodeModules) != 0 {
		nodeConfig.HTTPModules = append(nodeConfig.HTTPModules , nodeModules...)
	}

	//bootnode
	urls := make([]string , 0)
	if ctx.GlobalIsSet(utils.BootNodeUrlFlag.Name) {
		urls =  strings.Split(ctx.GlobalString(utils.BootNodeUrlFlag.Name) , ",")
	}

	for _ , url := range urls {
		n , err := discover.ParseNode(url)
		if err != nil {
			logger.Error("Bootstrap URL invalide encode:" , url , "err:" , err)
		}else{
			nodeConfig.P2P.BootstrapNodes = append(nodeConfig.P2P.BootstrapNodes , n)
		}
	}


	return &mjoy.MjoydConfig{
		Mjoy:mjoy.Config{
			Genesis:genesis.DefaultGenesisBlock(),
			NetworkId:uint64(params.WorkingChainId),
			Coinbase:types.Address{},
			TxPool:txprocessor.DefaultTxPoolConfig,
			StartBlockproducerAtStart:ctx.GlobalBool(utils.StartBlockproducerFlag.Name),
		},
		Node:nodeConfig,
	}
}

func createMjoyNode(ctx *cli.Context)(*node.Node){
	c := config.GetConfigInstance()
	c.SetPath(ctx.GlobalString(utils.ConfigFileFlag.Name))
	//mjoyMjoydConfig:=createMjoydCfg(ctx)

	stack , err := node.New()
	if err != nil{
		panic(fmt.Sprintf("node.New Wrong:%v", err))
	}
	stack.Register(func(ctx *node.ServiceContext)(node.Service , error){

		fullNode , err := mjoy.New(ctx)
		if err != nil{
			panic(fmt.Sprintf("mjoy.New Full Node:%v" , err))

		}
		logger.Debug("call the Constructor................")
		return fullNode , nil
	})

	return stack
}

func startMjoyNode(node *node.Node){

	if node == nil {
		logger.Critical("input node = nil")
		return
	}
	//start node
	if err := node.Start();err != nil {
		logger.Critical("Error starting protocol stack:",err)
	}

	go func(){
		sigc := make(chan os.Signal,1)
		signal.Notify(sigc , os.Interrupt)
		defer signal.Stop(sigc)

		<-sigc
		logger.Info("Got interrupt,shutting down...")
		go node.Stop()

		for i := 10;i > 0; i--{
			<-sigc
			if i > 1{
				logger.Warn("Already shutting down, interrupt more to panic.","times",i-1)
			}
		}
	}()

	node.Wait()

}

































