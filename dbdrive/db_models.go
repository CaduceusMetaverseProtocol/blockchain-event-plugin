package dbdrive

import (
	"blockchain-event-plugin/logger"
	"database/sql"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"strings"
	"sync"
	"time"
)

type Logs struct {
	Address     string   `json:"address" description:"address of the contract that generated the event"`
	Topics      []string `json:"topics" description:"list of topics provided by the contract"`
	Data        string   `json:"data" description:"supplied by the contract, usually ABI-encoded"`
	BlockNumber string   `json:"blockNumber" description:"block in which the transaction was included"`
	TxHash      string   `json:"transactionHash" description:"hash of the transaction"`
	TxIndex     string   `json:"transactionIndex" description:"index of the transaction in the block"`
	BlockHash   string   `json:"blockHash" description:"hash of the block in which the transaction was included"`
	LogIndex    string   `json:"logIndex" description:"index of the log in the block"`
	Removed     bool     `json:"removed" description:"The Removed field is true if this log was reverted due to a chain reorganisation"`
}

type BlockBloom struct {
	BlockNumber int64  `json:"blockNumber" description:"block in which the transaction was included"`
	BlockHash   string `json:"blockHash" description:"hash of the block in which the transaction was included"`
	Bloom       string `json:"bloom" description:"block bloom"`
}

// GetBloomByBlockNumber
func GetBloomByBlockNumber(blockNum int64) (bloom string, err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("GetLogs mysql error: ", r)
		}
	}()

	var rows *sql.Rows
	sql := "SELECT bloom FROM block_bloom WHERE block_number = ? limit 1"
	rows, err = DB.Query(sql, blockNum)

	defer rows.Close()
	if err != nil {
		panic(err)
	}
	// 数据处理
	for rows.Next() {
		rows.Scan(&bloom)
	}

	if err := rows.Err(); err != nil {
		panic(err)
	}

	return bloom, nil
}

// GetBlockNumAndBloomByBlockHash
func GetBlockNumAndBloomByBlockHash(blockHash string) (bloom BlockBloom, err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("GetBlockNumAndBloomByBlockHash mysql error: ", r)
		}
	}()

	var rows *sql.Rows
	sql := "SELECT block_number,bloom FROM block_bloom WHERE block_hash = ? limit 1"
	rows, err = DB.Query(sql, blockHash)

	defer rows.Close()
	if err != nil {
		panic(err)
	}
	// 数据处理
	for rows.Next() {
		rows.Scan(&bloom.BlockNumber, &bloom.Bloom)
	}

	if err := rows.Err(); err != nil {
		panic(err)
	}

	return bloom, nil
}

// GetLogsByBlockNum
func GetLogsByBlockNumber(blockNumber int64) (logs []Logs, err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("GetLogs mysql error: ", r)
		}
	}()

	var rows *sql.Rows
	sql := "SELECT address,topics,`data`,block_number,tx_hash,tx_index,block_hash,log_index,removed FROM logs WHERE block_number = ? "
	rows, err = DB.Query(sql, blockNumber)

	defer rows.Close()
	if err != nil {
		panic(err)
	}
	var log Logs
	// 数据处理
	for rows.Next() {
		var topic string
		var blockNumber int
		var address string
		rows.Scan(&address, &topic, &log.Data, &blockNumber, &log.TxHash, &log.TxIndex, &log.BlockHash, &log.LogIndex, &log.Removed)
		//处理address大小写
		log.Address = strings.ToLower(address)
		//处理Topics
		log.Topics = strings.Split(topic, ",")
		//处理block_number
		log.BlockNumber = toHex(blockNumber)
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		panic(err)
	}

	return logs, nil
}

// GetBlockNumber
func GetBlockHeight() (blockHeight int64, err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("GetLogs mysql error: ", r)
		}
	}()

	var rows *sql.Rows
	sql := "SELECT block_number FROM block_bloom order by block_number desc limit 1"
	rows, err = DB.Query(sql)

	defer rows.Close()
	if err != nil {
		panic(err)
	}
	// 数据处理
	for rows.Next() {
		rows.Scan(&blockHeight)
	}

	if err := rows.Err(); err != nil {
		panic(err)
	}

	return blockHeight, nil
}

func toHex(ten int) string {
	m := 0
	hex := make([]int, 0)
	for {
		m = ten % 16
		ten = ten / 16
		if ten == 0 {
			hex = append(hex, m)
			break
		}
		hex = append(hex, m)
	}
	hexStr := []string{}
	for i := len(hex) - 1; i >= 0; i-- {
		if hex[i] >= 10 {
			hexStr = append(hexStr, fmt.Sprintf("%c", 'a'+hex[i]-10))
		} else {
			hexStr = append(hexStr, fmt.Sprintf("%d", hex[i]))
		}
	}
	return "0x" + strings.Join(hexStr, "")

}

// save Logs
func SaveLogs(logs []ethtypes.Log) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("SaveLogs mysql error: ", r)
		}
	}()

	for index, value := range logs {
		if index < len(logs) {
			//处理topics
			topics := ""
			for i := 0; i < len(value.Topics); i++ {
				if i == len(value.Topics)-1 {
					topics += value.Topics[i].String()
				} else {
					topics += value.Topics[i].String() + ","
				}
			}
			_, err := Insert("INSERT INTO `logs`(`id`,`address`,`topics`,`data`,`block_number`,`tx_hash`,`tx_index`,`block_hash`,`log_index`,`removed`) values (?,?,?,?,?,?,?,?,?,?)",
				getSnowflakeId(), value.Address.String(), topics, "0x"+fmt.Sprintf("%x", value.Data), value.BlockNumber,
				value.TxHash.String(), hexutil.Uint64(value.TxIndex).String(), value.BlockHash.String(), hexutil.Uint64(value.Index).String(), fmt.Sprint(value.Removed))
			if err != nil {
				fmt.Println("SaveLogs Insert error: ", err)
			}
		}
	}
}

// Save block bloom
func SaveBloom(blockeHeight int64, blockHash, bloom string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("SaveLogs mysql error: ", r)
		}
	}()

	_, err := Insert("INSERT INTO `block_bloom`(`id`,`block_number`,`block_hash`,`bloom`) values (?,?,?,?)",
		getSnowflakeId(), blockeHeight, blockHash, bloom)
	if err != nil {
		fmt.Println("SaveBloom Insert error: ", err)
	}
}

// id生成器
var (
	machineID     int64 // 机器 id 占10位, 十进制范围是 [ 0, 1023 ]
	sn            int64 // 序列号占 12 位,十进制范围是 [ 0, 4095 ]
	lastTimeStamp int64 // 上次的时间戳(毫秒级), 1秒=1000毫秒, 1毫秒=1000微秒,1微秒=1000纳秒
	mu            sync.Mutex
)

func getSnowflakeId() int64 {
	mu.Lock()
	defer mu.Unlock()
	return getSnowflakeIdProcess()
}

func getSnowflakeIdProcess() int64 {
	curTimeStamp := time.Now().UnixNano() / 1000
	// 同一毫秒
	if curTimeStamp == lastTimeStamp {
		// 序列号占 12 位,十进制范围是 [ 0, 4095 ]
		if sn > 4095 {
			time.Sleep(time.Microsecond)
			curTimeStamp = time.Now().UnixNano() / 1000
			sn = 0
		}
	} else {
		sn = 0
	}
	sn++

	lastTimeStamp = curTimeStamp
	// 取 64 位的二进制数 0000000000 0000000000 0000000000 0001111111111 1111111111 1111111111  1 ( 这里共 41 个 1 )和时间戳进行并操作
	// 并结果( 右数 )第 42 位必然是 0,  低 41 位也就是时间戳的低 41 位
	rightBinValue := curTimeStamp & 0x1FFFFFFFFFF
	// 机器 id 占用10位空间,序列号占用12位空间,所以左移 22 位; 经过上面的并操作,左移后的第 1 位,必然是 0
	rightBinValue <<= 22
	id := rightBinValue | machineID | sn
	return id
}
