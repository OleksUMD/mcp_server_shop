package jsonrpc

const (
	ErrorParseError     = -32700
	ErrorInvalidRequest = -32600
	ErrorMethodNotFound = -32601
	ErrorInvalidParams  = -32602
	ErrorInternal       = -32603
)

func NewError(code int, message string, data any) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

func NewParseError(data any) *Error {
	return NewError(ErrorParseError, "Parse error", data)
}

func NewMethodNotFoundError(method string) *Error {
	return NewError(ErrorMethodNotFound, "Method not found: "+method, nil)
}

func NewInvalidParamsError(message string) *Error {
	return NewError(ErrorMethodNotFound, message, nil)
}

func NewInternalError(err error) *Error {
	return NewError(ErrorInternal, "Internal error", err.Error())
}
