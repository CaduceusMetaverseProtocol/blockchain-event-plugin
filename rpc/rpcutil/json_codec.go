package rpcutil

import (
	"bytes"
	"encoding/json"
	"github.com/ethereum/go-ethereum/eth/filters"
	"strings"
)

type jsonCodec struct {
	decBuf *bytes.Buffer
	dec    *json.Decoder
	encBuf *bytes.Buffer
	enc    *json.Encoder
}

// NewJSONCodec
func NewJSONCodec() *jsonCodec {
	decbuf := bytes.NewBuffer(nil)
	encbuf := bytes.NewBuffer(nil)
	c := &jsonCodec{
		decBuf: decbuf,
		dec:    json.NewDecoder(decbuf),
		encBuf: encbuf,
		enc:    json.NewEncoder(encbuf),
	}
	c.dec.DisallowUnknownFields()
	return c
}

func (j *jsonCodec) encode(argv interface{}) ([]byte, error) {
	return json.Marshal(argv)
}

func (j *jsonCodec) decode(data []byte, out interface{}) error {
	d := json.NewDecoder(strings.NewReader(string(data)))
	d.UseNumber()
	return d.Decode(&out)
}

func (j *jsonCodec) NewResponse(reply interface{}, err *jsonError) *jsonResponse {
	if err != nil {
		return &jsonResponse{
			Err:     err,
			Version: "2.0",
		}
	}

	jsonRes := new(jsonResponse)
	jsonB, _ := json.Marshal(reply)
	json.Unmarshal(jsonB, &jsonRes)
	if jsonRes.Err != nil {
		return &jsonResponse{
			Version: "2.0",
			Err:     jsonRes.Err,
		}
	}
	if jsonRes.Result != nil {
		return &jsonResponse{
			Version: "2.0",
			Result:  jsonRes.Result,
		}
	}
	response := &jsonResponse{
		Version: "2.0",
		Result:  reply,
	}
	return response
}

func (j *jsonCodec) NewRequest(id int, method string, argv []filters.FilterCriteria) *jsonRequest {
	req := &jsonRequest{
		ID:      id,
		Mthd:    method,
		Args:    argv,
		Version: "2.0",
	}
	return req
}

func (j *jsonCodec) ReadResponse(data []byte) (resps []*jsonResponse, err error) {
	jsonResps := make([]*jsonResponse, 0)
	if err := j.decode(data, &jsonResps); err != nil {
		resp := new(jsonResponse)
		if err := j.decode(data, resp); err != nil {
			return nil, err
		}
		resps = append(resps, resp)
		return resps, nil
	}

	for _, jsonResp := range jsonResps {
		resps = append(resps, jsonResp)
	}

	return resps, nil
}

func (j *jsonCodec) ReadRequest(data []byte) (reqs *jsonRequest, err error) {
	req := new(jsonRequest)
	if err := j.decode(data, req); err != nil {
		return nil, err
	}
	return req, nil
}

func (j *jsonCodec) ReadRequestBody(data []byte, rcvr interface{}) error {
	return j.decode(data, rcvr)
}

func (j *jsonCodec) ReadResponseBody(data []byte, rcvr interface{}) error {
	return j.decode(data, rcvr)
}

func (j *jsonCodec) EncodeRequests(v interface{}) ([]byte, error) {
	return j.encode(v)
}

func (j *jsonCodec) EncodeResponses(v interface{}) ([]byte, error) {
	return j.encode(v)
}
