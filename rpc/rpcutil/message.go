package rpcutil

// Request interface contains necessary methods
type Request interface {
	Method() string
	Params() []byte
	Ident() string
}

// Response interface contains necessary methods
type Response interface {
	Error() error
	ErrCode() int
	Reply() []byte
	SetReqIdent(ident string)
}
