package rpcutil

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/eth/filters"
)

type jsonError struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

func (j *jsonError) Error(code int64, errMsg, errData string) *jsonError {
	return &jsonError{
		Code:    code,
		Message: errMsg,
		Data:    errData,
	}
}

// jsonRequest is jsonCodec response data struct
// and implement the inerface named 'rpc.Request'
type jsonRequest struct {
	ID      int                      `json:"id"`
	Mthd    string                   `json:"method"`
	Args    []filters.FilterCriteria `json:"params"`
	Version string                   `json:"jsonrpc"`
}

func (j *jsonRequest) Ident() int     { return j.ID }
func (j *jsonRequest) Method() string { return j.Mthd }
func (j *jsonRequest) Params() []byte {
	byts, err := json.Marshal(j.Args)
	if err != nil {
		panic(err)
	}
	return byts
}

// jsonResponse is jsonCodec response data struct
// and implement the inerface named 'rpc.Response'
type jsonResponse struct {
	ID      int         `json:"id"`
	Version string      `json:"jsonrpc"`
	Err     *jsonError  `json:"error,omitempty"`
	Result  interface{} `json:"result,omitempty"`
}

func (j *jsonResponse) SetReqIdent(ident int) { j.ID = ident }
func (j *jsonResponse) Error() *jsonError     { return j.Err }
func (j *jsonResponse) Reply() []byte {
	byts, err := json.Marshal(j)
	if err != nil {
		panic(err)
	}
	return byts
}
func (j *jsonResponse) ErrCode() int64 {
	if j.Err != nil {
		return j.Err.Code
	}
	return -1
}
