package rpcutil

import (
	"errors"
	"log"
	"reflect"
	"unicode"
	"unicode/utf8"
)

var (
	typeOfError = reflect.TypeOf((*error)(nil)).Elem()
)

type service struct {
	name   string
	rcvr   reflect.Value
	typ    reflect.Type
	method map[string]*methodType
}

type methodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
}

func (s *Server) Register(rcvr interface{}) error {
	srvic := new(service)
	srvic.typ = reflect.TypeOf(rcvr)
	srvic.rcvr = reflect.ValueOf(rcvr)
	sname := reflect.Indirect(srvic.rcvr).Type().Name()
	if sname == "" {
		return errors.New("rpc.Register: no service name for type " + srvic.typ.String())
	}

	if !isExported(sname) {
		return errors.New("rpc.Register: type " + sname + " is not exported")
	}
	srvic.name = sname
	srvic.method = suitableMethods(srvic.typ)

	if _, dup := s.m.LoadOrStore(sname, srvic); dup {
		return errors.New("rpc: service already defined: " + sname)
	}
	return nil
}

// Before Call must parse and decode param into reflect.Value
// after Call must encode and response
func (s *Server) call(req *jsonRequest) (reply *jsonResponse) {
	serviceName, methodName := parseFromRPCMethod(req.Method())
	// method existed or not
	svci, ok := s.m.Load(serviceName)
	if !ok {
		err := errors.New("rpc: can't find service " + serviceName)
		jsonErr := new(jsonError).Error(-32601, "Method not found", err.Error())
		reply = s.codec.NewResponse(nil, jsonErr)
		reply.SetReqIdent(req.Ident())
		return
	}

	svc := svci.(*service)
	mtype := svc.method[methodName]
	if mtype == nil {
		err := errors.New("rpc: can't find method " + req.Method())
		jsonErr := new(jsonError).Error(-32601, "Method not found", err.Error())
		reply = s.codec.NewResponse(nil, jsonErr)
		reply.SetReqIdent(req.Ident())
		return
	}

	argv := reflect.ValueOf(req.Args[0])
	var replyv reflect.Value
	replyv = reflect.New(mtype.ReplyType.Elem())

	if err := svc.call(mtype, argv, replyv); err != nil {
		jsonErr := new(jsonError).Error(-32603, "Internal error", err.Error())
		reply = s.codec.NewResponse(nil, jsonErr)
		reply.SetReqIdent(req.Ident())
	} else {
		reply = s.codec.NewResponse(replyv.Interface(), nil)
		reply.SetReqIdent(req.Ident())
	}
	return
}

func (s *service) call(mtype *methodType, argv reflect.Value, replyv reflect.Value) error {
	function := mtype.method.Func
	returnValues := function.Call([]reflect.Value{s.rcvr, argv, replyv})
	if i := returnValues[0].Interface(); i != nil {
		return i.(error)
	}
	return nil
}

func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

func suitableMethods(typ reflect.Type) map[string]*methodType {
	methods := make(map[string]*methodType)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		if mt := suitableMethod(method); mt != nil {
			methods[method.Name] = mt
		}
	}
	return methods
}

func suitableMethod(method reflect.Method) *methodType {
	mtype := method.Type
	mname := method.Name

	// Method must be exported.
	if method.PkgPath != "" {
		return nil
	}
	// Method needs three ins: receiver, *args, *reply.
	if mtype.NumIn() != 3 {
		log.Printf("rpc.Register: method %q has %d input parameters; needs exactly three\n", mname, mtype.NumIn())
		return nil
	}
	// First arg need not be a pointer.
	argType := mtype.In(1)
	if !isExportedOrBuiltinType(argType) {
		log.Printf("rpc.Register: argument type of method %q is not exported: %q\n", mname, argType)
		return nil
	}
	// Second arg must be a pointer.
	replyType := mtype.In(2)
	if replyType.Kind() != reflect.Ptr {
		log.Printf("rpc.Register: reply type of method %q is not a pointer: %q\n", mname, replyType)
		return nil
	}
	// Reply type must be exported.
	if !isExportedOrBuiltinType(replyType) {
		log.Printf("rpc.Register: reply type of method %q is not exported: %q\n", mname, replyType)
		return nil
	}
	// Method needs one out.
	if mtype.NumOut() != 1 {
		log.Printf("rpc.Register: method %q has %d output parameters; needs exactly one\n", mname, mtype.NumOut())
		return nil
	}
	// The return type of the method must be error.
	if returnType := mtype.Out(0); returnType != typeOfError {
		log.Printf("rpc.Register: return type of method %q is %q, must be error\n", mname, returnType)
		return nil
	}
	return &methodType{method: method, ArgType: argType, ReplyType: replyType}
}

// Is this type exported or a builtin?
func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Slice {
		t = t.Elem()
	}
	return isExported(t.Name()) || t.PkgPath() == ""
}
