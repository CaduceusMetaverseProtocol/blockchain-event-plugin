package types

import "github.com/ethereum/go-ethereum/common/hexutil"

type Error struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

type Response struct {
	ID      string      `json:"id"`
	Version string      `json:"jsonrpc"`
	Error   interface{} `json:"error,omitempty"`
	Result  interface{} `json:"result,omitempty"`
}

type Block struct {
	Number    hexutil.Uint64
	Hash      string
	LogsBloom string
}
