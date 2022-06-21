package rpcutil

import (
	"blockchain-event-plugin/logger"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

type Server struct {
	status int
	l      sync.Mutex
	m      sync.Map    // map[string]*service
	codec  ServerCodec // codec to read request and writeResponse
}

func NewServer() *Server {
	codec := NewJSONCodec()
	return &Server{
		status: 0,
		codec:  codec,
	}
}

// ListenHTTPServe open http support can serve http request
func (s *Server) ListenHTTPServe(addr string) {
	if err := http.ListenAndServe(
		addr,
		http.TimeoutHandler(s, 300*time.Second, "Network request timeout"),
	); err != nil {
		logger.Fatal("RPC server over HTTP is error: ", err)
	}
}

// ServeHTTP handle request over HTTP,
// it also implement the interface of http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if err, ok := recover().(error); ok && err != nil {
			logger.Error("ServeHTTP recover:", "err", err)
			debug.PrintStack()
		}
	}()

	if s.GetState() == 1 {
		jsonErr := new(jsonError).Error(-32603, "Internal error", "Node channel closed")
		resp := s.codec.NewResponse(nil, jsonErr)
		byts, _ := s.codec.EncodeResponses(resp)
		String(w, http.StatusUnauthorized, byts)
		return
	}

	var (
		data []byte
		err  error
	)
	switch req.Method {
	case http.MethodPost:
		if data, err = ioutil.ReadAll(req.Body); err != nil {
			jsonErr := new(jsonError).Error(-32603, "Internal error", err.Error())
			resp := s.codec.NewResponse(nil, jsonErr)
			byts, err := s.codec.EncodeResponses(resp)
			logger.Warn("S.codec.EncodeResponses：", "err", err)
			String(w, http.StatusOK, byts)
			return
		}
		defer req.Body.Close()
	default:
		err := errors.New("method not allowed: " + req.Method)
		jsonErr := new(jsonError).Error(-32601, "Method not found", err.Error())
		resp := s.codec.NewResponse(nil, jsonErr)
		byts, err := s.codec.EncodeResponses(resp)
		logger.Warn("S.codec.EncodeResponses：", "err", err)
		String(w, http.StatusOK, byts)
		return
	}

	rpcReq, err := s.codec.ReadRequest(data)
	if err != nil {
		jsonErr := new(jsonError).Error(-32600, "Invalid Request", err.Error())
		resp := s.codec.NewResponse(nil, jsonErr)
		byts, err := s.codec.EncodeResponses(resp)
		if err != nil {
			logger.Error("S.codec.EncodeResponses：", "err", err)
		}
		String(w, http.StatusOK, byts)
		return
	}
	resps := s.call(rpcReq)
	byts, _ := s.codec.EncodeResponses(resps)
	String(w, http.StatusOK, byts)
	return
}

func (s *Server) SetState(sta int) {
	s.l.Lock()
	s.status = sta
	s.l.Unlock()
}

func (s *Server) GetState() int {
	s.l.Lock()
	defer s.l.Unlock()
	return s.status
}

func parseFromRPCMethod(reqMethod string) (serviceName, methodName string) {
	if strings.Count(reqMethod, "_") != 1 {
		return "", ""
	}
	serviceName, methodName = StrFirstToUpper(reqMethod)
	serviceName = "PublicRPCAPI"
	return serviceName, methodName
}

// String
func String(w http.ResponseWriter, statusCode int, byts []byte) error {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "text/plain")
	_, err := io.WriteString(w, string(byts))
	return err
}

// String first character to upper.  abc -> Abc
func StrFirstToUpper(str string) (string, string) {
	temp := strings.Split(str, "_")
	var upperStr string
	for y := 0; y < len(temp); y++ {
		vv := []rune(temp[y])
		if y != 0 {
			for i := 0; i < len(vv); i++ {
				if i == 0 {
					vv[i] -= 32
					upperStr += string(vv[i]) // + string(vv[i+1])
				} else {
					upperStr += string(vv[i])
				}
			}
		}
	}
	return temp[0], upperStr
}
