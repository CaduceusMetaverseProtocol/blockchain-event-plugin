package types

var version = "2.0"

func ErrorMsg(code int64, message, msgData string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Data:    msgData,
	}
}

func Responses(id string, errmsg *Error, result interface{}) *Response {
	if errmsg != nil {
		return &Response{
			ID:      id,
			Version: version,
			Error:   errmsg,
		}
	}
	return &Response{
		ID:      id,
		Version: version,
		Result:  result,
	}
}
