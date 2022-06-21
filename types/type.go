package types

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
