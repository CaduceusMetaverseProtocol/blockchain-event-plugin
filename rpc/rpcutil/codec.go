package rpcutil

type ClientCodec interface {
	// generate a single NewRequest with needed params
	NewRequest(id, method string, argv []string) *jsonRequest
	// EncodeRequests .
	EncodeRequests(v interface{}) ([]byte, error)
	// parse encoded data into a Response
	ReadResponse(data []byte) ([]*jsonResponse, error)
	// ReadResponseBody .
	ReadResponseBody(respBody []byte, rcvr interface{}) error
}

type ServerCodec interface {
	// parse encoded data into a Request
	ReadRequest(data []byte) (*jsonRequest, error)
	// ReadRequestBody parse params
	ReadRequestBody(reqBody []byte, rcvr interface{}) error
	// generate a single Response with needed params
	NewResponse(replyv interface{}, err *jsonError) *jsonResponse
	// EncodeResponses .
	EncodeResponses(v interface{}) ([]byte, error)
}
