package types

var (
	InvalidRequest = &Error{Code: -32600, Message: "Invalid Request"}
	MethodNotFound = &Error{Code: -32601, Message: "Method not found"}
	SystemError    = &Error{Code: -32603, Message: "Internal error"}
	InvalidParams  = &Error{Code: -32000, Message: "Invalid params"}
)
