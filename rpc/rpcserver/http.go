package rpcserver

import (
	"blockchain-event-plugin/dbdrive"
	"blockchain-event-plugin/logger"
	"blockchain-event-plugin/rpc/filter"
	"blockchain-event-plugin/rpc/rpcutil"
	"blockchain-event-plugin/types"
	"github.com/ethereum/go-ethereum/eth/filters"
	"os"
	"time"
)

// 启动HTTP RPC
func StartRPC(addr string) {
	server := rpcutil.NewServer()
	err := server.Register(new(PublicRPCAPI))
	if err != nil {
		logger.Error("StartRPC Register err", err)
	}

	go func() {
		// 监听退出信号
		s := <-make(chan os.Signal)
		// 关闭网络服务
		server.SetState(1)
		logger.Debug("[sig] exit signal capture", "signal", s)
	}()

	logger.Info("[sys] Listen HTTP RPC on", addr)
	server.ListenHTTPServe(addr)
}

type PublicRPCAPI struct {
}

// GetLogs
func (i *PublicRPCAPI) GetLogs(crit filters.FilterCriteria, reply *interface{}) error {

	start := time.Now()

	if len(crit.Addresses) == 0 && len(crit.Topics) == 0 && crit.BlockHash == nil {
		types.InvalidParams.Data = "Parameters is empty"
		*reply = types.Responses("000000", types.InvalidParams, nil)
		return nil
	}

	publicFilterAPI := filter.NewPublicAPI()
	logs, err := publicFilterAPI.HandleGetLogs(crit)
	if err != nil {
		logger.Error("GetLogs error", "args", crit, "err", err)
		types.InvalidParams.Data = err.Error()
		*reply = types.Responses("000000", types.InvalidParams, nil)
		return nil
	}
	if logs == nil || len(logs) == 0 {
		*reply = []dbdrive.Logs{}
	} else {
		*reply = logs
	}
	logger.Info("eth_getLogs end!", "startTime:", start.UnixNano()/1000/1000, "cost:", time.Now().Sub(start).Milliseconds(), "ms, params:", crit)
	return nil
}
