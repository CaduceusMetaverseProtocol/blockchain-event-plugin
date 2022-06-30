package main

import (
	"blockchain-event-plugin/logger"
	"blockchain-event-plugin/rpc/rpcserver"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 加载日志配置
	logger.SetLogger("config/log.json")
	Run()
}

// Run 开始运行
func Run() {

	// 监听中断信号
	signal.Notify(make(chan os.Signal), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// 服务开启
	rpcPort := viper.GetString("rpc.port")

	if rpcPort == "" {
		rpcPort = os.Getenv("RPC_PORT")
		logger.Info("Command line get RPC_PORT:", rpcPort)
	}
	go rpcserver.StartRPC(":" + rpcPort)

	logger.Info("[sys] CMP service start successful: ", "time", time.Now().UTC())

	<-make(chan struct{})
}
