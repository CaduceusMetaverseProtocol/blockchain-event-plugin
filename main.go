package main

import (
	"blockchain-event-plugin/logger"
	"blockchain-event-plugin/rpc/rpcserver"
	"blockchain-event-plugin/setting"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 加载日志配置
	logger.SetLogger(setting.GetString("logger_jsonFile"))
	Run()
}

// Run 开始运行
func Run() {

	// 监听中断信号
	signal.Notify(make(chan os.Signal), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// 服务开启
	go rpcserver.StartRPC(setting.GetString("rpc.port"))

	logger.Info("[sys] CMP service start successful: ", "time", time.Now().UTC())

	<-make(chan struct{})
}
