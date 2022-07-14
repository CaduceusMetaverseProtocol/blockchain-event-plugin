package rpcserver

import (
	"blockchain-event-plugin/dbdrive"
	"blockchain-event-plugin/logger"
	"blockchain-event-plugin/rpc/filter"
	"blockchain-event-plugin/rpc/rpcutil"
	"blockchain-event-plugin/types"
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/eth/filters"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"math/big"
	"os"
	"strconv"
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

//Save logs
func (i *PublicRPCAPI) SyncBlockAndLogs(crit filters.FilterCriteria, reply *interface{}) error {

	url := os.Getenv("SyncRpcAddr")
	//url := "https://mainnet.block.caduceus.foundation"

	//-------------------- 从链上查指定高度的区块信息并存储 --------------------
	rpcclient, err := rpc.Dial(url)
	if rpcclient == nil {
		logger.Error("SyncBlockAndLogs rpcclient dial url failed.", "args:", crit, "err:", err)
		types.SystemError.Data = err.Error()
		*reply = types.Responses("000000", types.SystemError, nil)
		return nil
	}

	fromBlockStr := crit.FromBlock.String()
	toBlockStr := crit.ToBlock.String()
	fromBlockInt, err := strconv.ParseInt(fromBlockStr, 10, 64)
	if err != nil {
		logger.Error("SyncBlockAndLogs Parse FromBlock error.", "args:", crit, "err:", err)
		types.InvalidParams.Data = err.Error()
		*reply = types.Responses("000000", types.InvalidParams, nil)
		return nil
	}
	toBlockInt, err := strconv.ParseInt(toBlockStr, 10, 64)
	if err != nil {
		logger.Error("SyncBlockAndLogs Parse ToBlock error.", "args:", crit, "err:", err)
		types.InvalidParams.Data = err.Error()
		*reply = types.Responses("000000", types.InvalidParams, nil)
		return nil
	}

	var raw json.RawMessage
	for i := fromBlockInt; i <= toBlockInt; i++ {
		err := rpcclient.Call(&raw, "eth_getBlockByNumber", hexutil.EncodeBig(big.NewInt(i)), false)
		if err != nil {
			logger.Error("SyncBlockAndLogs rpcclient.Call error.", "args:", crit, "err:", err)
			types.SystemError.Data = err.Error()
			*reply = types.Responses("000000", types.SystemError, nil)
			return nil
		}
		//处理数据
		var block types.Block
		json.Unmarshal(raw, &block)

		//save block_blomm
		dbdrive.SaveBloom(int64(block.Number), block.Hash, block.LogsBloom)
	}

	//---------------- 从链上根据区块高度查询logs并存储 --------------------
	client, err := ethclient.Dial(url)
	if err != nil {
		logger.Error("SyncBlockAndLogs ethclient dial url failed.", "args:", crit, "err:", err)
		types.SystemError.Data = err.Error()
		*reply = types.Responses("000000", types.SystemError, nil)
		return nil
	}
	ctx := context.Background()
	queryParam := ethereum.FilterQuery{
		FromBlock: crit.FromBlock,
		ToBlock:   crit.ToBlock,
	}
	ethlogs, err := client.FilterLogs(ctx, queryParam)
	if err != nil {
		logger.Error("SyncBlockAndLogs FilterLogs failed.", "args:", crit, "err:", err)
		types.SystemError.Data = err.Error()
		*reply = types.Responses("000000", types.SystemError, nil)
		return nil
	}

	//save logs
	dbdrive.SaveLogs(ethlogs)

	*reply = "Sync successful!"
	logger.Info("Sync successful!", "fromBlock:", crit.FromBlock, "; toBlock:", crit.ToBlock)
	return nil
}
