package filter

import (
	"blockchain-event-plugin/dbdrive"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/filters"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/pkg/errors"
	"sync"
	"time"
)

// consider a filter inactive if it has not been polled for within deadline
var deadline = 10 * time.Minute

const (
	DefaultLogsCap       int32 = 10000
	DefaultBlockRangeCap int32 = 10000
)

// filter is a helper struct that holds meta information over the filter type
// and associated subscription in the event system.
type filter struct {
	typ      filters.Type
	deadline *time.Timer // filter is inactive when deadline triggers
	hashes   []common.Hash
	crit     filters.FilterCriteria
	logs     []*ethtypes.Log
}

// Backend defines the methods requided by the PublicFilterAPI backend
type Backend interface {
	GetLogs(blockHash common.Hash) ([][]*ethtypes.Log, error)
	RPCLogsCap() int32
	RPCBlockRangeCap() int32
}

// PublicFilterAPI offers support to create and manage filter. This will allow external clients to retrieve various
// information related to the Ethereum protocol such as blocks, transactions and logs.
type PublicFilterAPI struct {
	backend   Backend
	filtersMu sync.Mutex
	filters   map[rpc.ID]*filter
}

// NewPublicAPI returns a new PublicFilterAPI instance.
func NewPublicAPI() *PublicFilterAPI {
	api := &PublicFilterAPI{
		filters: make(map[rpc.ID]*filter),
	}

	go api.timeoutLoop()

	return api
}

// timeoutLoop runs every 5 minutes and deletes filters that have not been recently used.
// Tt is started when the api is created.
func (api *PublicFilterAPI) timeoutLoop() {
	ticker := time.NewTicker(deadline)
	defer ticker.Stop()

	for {
		<-ticker.C
		api.filtersMu.Lock()
		for id, f := range api.filters {
			select {
			case <-f.deadline.C:
				delete(api.filters, id)
			default:
				continue
			}
		}
		api.filtersMu.Unlock()
	}
}

func (api *PublicFilterAPI) HandleGetLogs(crit filters.FilterCriteria) ([]dbdrive.Logs, error) {
	var filter *Filter
	if crit.BlockHash != nil {
		// Block filter requested, construct a single-shot filter
		filter = NewBlockFilter(api.backend, crit)
	} else {
		// Convert the RPC block numbers into internal representations
		begin, err := dbdrive.GetBlockHeight()
		if err != nil {
			return nil, errors.Wrap(err, "HandleGetLogs getBlockHeight for FromBlock error")
		}
		if crit.FromBlock != nil {
			begin = crit.FromBlock.Int64()
		}
		end, err := dbdrive.GetBlockHeight()
		if err != nil {
			return nil, errors.Wrap(err, "HandleGetLogs getBlockHeight for toBlock error")
		}
		if crit.ToBlock != nil {
			end = crit.ToBlock.Int64()
		}
		// Construct the range filter
		filter = NewRangeFilter(api.backend, begin, end, crit.Addresses, crit.Topics)
	}

	// Run the filter and return all the logs
	logs, err := filter.Logs(int(DefaultLogsCap), int64(DefaultBlockRangeCap))
	if err != nil {
		return nil, err
	}

	return returnLogs(logs), err
}
