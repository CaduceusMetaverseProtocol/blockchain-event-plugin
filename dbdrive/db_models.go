package dbdrive

import (
	"blockchain-event-plugin/logger"
	"database/sql"
	"fmt"
	"strings"
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
