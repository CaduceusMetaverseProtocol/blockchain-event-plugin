package filter

import (
	"blockchain-event-plugin/dbdrive"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"strconv"
	"strings"
)

// filterLogs creates a slice of logs matching the given criteria.
// [] -> anything
// [A] -> A in first position of log topics, anything after
// [null, B] -> anything in first position, B in second position
// [A, B] -> A in first position and B in second position
// [[A, B], [A, B]] -> A or B in first position, A or B in second position
func FilterLogs(logs []dbdrive.Logs, fromBlock, toBlock *big.Int, addresses []common.Address, topics [][]common.Hash) []dbdrive.Logs {
	var ret []dbdrive.Logs
Logs:
	for _, log := range logs {
		if fromBlock != nil && fromBlock.Int64() >= 0 && fromBlock.Uint64() > strToUint64(log.BlockNumber) {
			continue
		}
		if toBlock != nil && toBlock.Int64() >= 0 && toBlock.Uint64() < strToUint64(log.BlockNumber) {
			continue
		}
		if len(addresses) > 0 && !includes(addresses, log.Address) {
			continue
		}
		// If the to filtered topics is greater than the amount of topics in logs, skip.
		if len(topics) > len(log.Topics) {
			continue
		}
		for i, sub := range topics {
			match := len(sub) == 0 // empty rule set == wildcard
			for _, topic := range sub {
				if log.Topics[i] == topic.String() {
					match = true
					break
				}
			}
			if !match {
				continue Logs
			}
		}
		ret = append(ret, log)
	}
	return ret
}

func includes(addresses []common.Address, a string) bool {
	for _, addr := range addresses {
		if strings.ToLower(addr.String()) == a {
			return true
		}
	}
	return false
}

func bloomFilter(bloom ethtypes.Bloom, addresses []common.Address, topics [][]common.Hash) bool {
	if len(addresses) > 0 {
		var included bool
		for _, addr := range addresses {
			if ethtypes.BloomLookup(bloom, addr) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}

	for _, sub := range topics {
		included := len(sub) == 0 // empty rule set == wildcard
		for _, topic := range sub {
			if ethtypes.BloomLookup(bloom, topic) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}
	return true
}

// returnHashes is a helper that will return an empty hash array case the given hash array is nil,
// otherwise the given hashes array is returned.
func returnHashes(hashes []common.Hash) []common.Hash {
	if hashes == nil {
		return []common.Hash{}
	}
	return hashes
}

// returnLogs is a helper that will return an empty log array in case the given logs array is nil,
// otherwise the given logs array is returned.
func returnLogs(logs []dbdrive.Logs) []dbdrive.Logs {
	if logs == nil {
		return []dbdrive.Logs{}
	}
	return logs
}

func strToUint64(str string) uint64 {
	intNum, _ := strconv.Atoi(str)
	return uint64(intNum)
}
