package filter

import (
	"blockchain-event-plugin/dbdrive"
	"blockchain-event-plugin/logger"
	"encoding/binary"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth/filters"
	"github.com/pkg/errors"
	"math/big"
)

const (
	maxToOverhang = 600
)

// Filter can be used to retrieve and filter logs.
type Filter struct {
	backend  Backend
	criteria filters.FilterCriteria

	bloomFilters [][]BloomIV // Filter the system is matching for
}

// BloomIV represents the bit indexes and value inside the bloom filter that belong
// to some key.
type BloomIV struct {
	I [3]uint
	V [3]byte
}

func NewBlockFilter(backend Backend, criteria filters.FilterCriteria) *Filter {
	// Create a generic filter and convert it into a block filter
	return newFilter(backend, criteria, nil)
}

// newFilter returns a new Filter
func newFilter(backend Backend, criteria filters.FilterCriteria, bloomFilters [][]BloomIV) *Filter {
	return &Filter{
		backend:      backend,
		criteria:     criteria,
		bloomFilters: bloomFilters,
	}
}

// NewRangeFilter creates a new filter which uses a bloom filter on blocks to
// figure out whether a particular block is interesting or not.
func NewRangeFilter(backend Backend, begin, end int64, addresses []common.Address, topics [][]common.Hash) *Filter {
	// Flatten the address and topic filter clauses into a single bloombits filter system.
	// Since the bloombits are not positional, nil topics are permitted,
	// which get flattened into a nil byte slice.
	var filtersBz [][][]byte // nolint: prealloc
	if len(addresses) > 0 {
		filter := make([][]byte, len(addresses))
		for i, address := range addresses {
			filter[i] = address.Bytes()
		}
		filtersBz = append(filtersBz, filter)
	}

	for _, topicList := range topics {
		filter := make([][]byte, len(topicList))
		for i, topic := range topicList {
			filter[i] = topic.Bytes()
		}
		filtersBz = append(filtersBz, filter)
	}

	// Create a generic filter and convert it into a range filter
	criteria := filters.FilterCriteria{
		FromBlock: big.NewInt(begin),
		ToBlock:   big.NewInt(end),
		Addresses: addresses,
		Topics:    topics,
	}

	return newFilter(backend, criteria, createBloomFilters(filtersBz))
}

func createBloomFilters(filters [][][]byte) [][]BloomIV {
	bloomFilters := make([][]BloomIV, 0)
	for _, filter := range filters {
		// Gather the bit indexes of the filter rule, special casing the nil filter
		if len(filter) == 0 {
			continue
		}
		bloomIVs := make([]BloomIV, len(filter))

		// Transform the filter rules (the addresses and topics) to the bloom index and value arrays
		// So it can be used to compare with the bloom of the block header. If the rule has any nil
		// clauses. The rule will be ignored.
		for i, clause := range filter {
			if clause == nil {
				bloomIVs = nil
				break
			}

			iv, err := calcBloomIVs(clause)
			if err != nil {
				bloomIVs = nil
				logger.Error("calcBloomIVs error: %v", err)
				break
			}

			bloomIVs[i] = iv
		}
		// Accumulate the filter rules if no nil rule was within
		if bloomIVs != nil {
			bloomFilters = append(bloomFilters, bloomIVs)
		}
	}
	return bloomFilters
}

// calcBloomIVs returns BloomIV for the given data,
func calcBloomIVs(data []byte) (BloomIV, error) {
	hashbuf := make([]byte, 6)
	biv := BloomIV{}

	sha := crypto.NewKeccakState()
	sha.Reset()
	if _, err := sha.Write(data); err != nil {
		return BloomIV{}, err
	}
	if _, err := sha.Read(hashbuf); err != nil {
		return BloomIV{}, err
	}

	// The actual bits to flip
	biv.V[0] = byte(1 << (hashbuf[1] & 0x7))
	biv.V[1] = byte(1 << (hashbuf[3] & 0x7))
	biv.V[2] = byte(1 << (hashbuf[5] & 0x7))
	// The indices for the bytes to OR in
	biv.I[0] = ethtypes.BloomByteLength - uint((binary.BigEndian.Uint16(hashbuf)&0x7ff)>>3) - 1
	biv.I[1] = ethtypes.BloomByteLength - uint((binary.BigEndian.Uint16(hashbuf[2:])&0x7ff)>>3) - 1
	biv.I[2] = ethtypes.BloomByteLength - uint((binary.BigEndian.Uint16(hashbuf[4:])&0x7ff)>>3) - 1

	return biv, nil
}

// Logs searches the blockchain for matching log entries, returning all from the
// first block that contains matches, updating the start of the filter accordingly.
func (f *Filter) Logs(logLimit int, blockLimit int64) ([]dbdrive.Logs, error) {
	logs := []dbdrive.Logs{}
	var err error

	// If we're doing singleton block filtering, execute and return
	if f.criteria.BlockHash != nil && *f.criteria.BlockHash != (common.Hash{}) {
		// get bloom
		blockBloom, err := dbdrive.GetBlockNumAndBloomByBlockHash(f.criteria.BlockHash.String())
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch header by hash")
		}
		if &blockBloom == nil {
			return nil, errors.Errorf("unknown bloom %s", f.criteria.BlockHash.String())
		}

		//处理bloom
		byteBloom, err := hex.DecodeString(blockBloom.Bloom)
		bloom := ethtypes.BytesToBloom(byteBloom)

		return f.blockLogs(blockBloom.BlockNumber, bloom)
	}

	// Figure out the limits of the filter range
	blockHeight, err := dbdrive.GetBlockHeight()
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch block height")
	}

	if f.criteria.FromBlock.Int64() < 0 {
		f.criteria.FromBlock = big.NewInt(blockHeight)
	} else if f.criteria.FromBlock.Int64() == 0 {
		f.criteria.FromBlock = big.NewInt(1)
	}
	if f.criteria.ToBlock.Int64() < 0 {
		f.criteria.ToBlock = big.NewInt(blockHeight)
	} else if f.criteria.ToBlock.Int64() == 0 {
		f.criteria.ToBlock = big.NewInt(1)
	}

	if f.criteria.ToBlock.Int64()-f.criteria.FromBlock.Int64() > blockLimit {
		return nil, errors.Errorf("maximum [from, to] blocks distance: %d", blockLimit)
	}

	// check bounds
	if f.criteria.FromBlock.Int64() > blockHeight {
		return []dbdrive.Logs{}, nil
	} else if f.criteria.ToBlock.Int64() > blockHeight+maxToOverhang {
		f.criteria.ToBlock = big.NewInt(blockHeight + maxToOverhang)
	}

	from := f.criteria.FromBlock.Int64()
	to := f.criteria.ToBlock.Int64()

	for height := from; height <= to; height++ {
		// 根据区块高度获取bloom
		blockBloom, err := dbdrive.GetBloomByBlockNumber(height)
		if err != nil {
			return nil, err
		}
		if blockBloom == "" {
			logger.Debug("Block bloom not found or has no number")
			return nil, nil
		}
		logger.Debug("api logs", " get block bloom ", blockBloom)

		byteBloom, err := hex.DecodeString(blockBloom)
		bloom := ethtypes.BytesToBloom(byteBloom)
		filtered, err := f.blockLogs(height, bloom)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to fetch block by number %d", height)
		}
		logger.Debug("api logs ", "filtered logs ", filtered)

		// check logs limit
		if len(logs)+len(filtered) > logLimit {
			return nil, errors.Errorf("query returned more than %d results", logLimit)
		}
		logs = append(logs, filtered...)
	}
	return logs, nil
}

// blockLogs returns the logs matching the filter criteria within a single block.
func (f *Filter) blockLogs(height int64, bloom ethtypes.Bloom) ([]dbdrive.Logs, error) {
	if !bloomFilter(bloom, f.criteria.Addresses, f.criteria.Topics) {
		return []dbdrive.Logs{}, nil
	}

	var logsList []dbdrive.Logs
	logsList, err := dbdrive.GetLogsByBlockNumber(height)
	if err != nil {
		return []dbdrive.Logs{}, errors.Wrapf(err, "failed to fetch logs by block number %d", height)
	}

	unfiltered := make([]dbdrive.Logs, 0)
	for _, logs := range logsList {
		unfiltered = append(unfiltered, logs)
	}
	logs := FilterLogs(unfiltered, nil, nil, f.criteria.Addresses, f.criteria.Topics)
	if len(logs) == 0 {
		return []dbdrive.Logs{}, nil
	}
	return logs, nil
}
