package httpserver

import (
	"net/http"
	"strconv"
)

// Error ...
type Error struct {
	Code    int               `json:"code,omitempty"`
	Message string            `json:"message,omitempty"`
	Data    map[string]string `json:"data,omitempty"`
}

func (e *Error) GetCode() int {
	if e == nil {
		return http.StatusOK
	}
	return e.Code
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return "code :" + strconv.Itoa(e.Code) + " msg: " + e.Message
}

func buildError(err error, code int) *Error {
	return &Error{
		Message: err.Error(),
		Code:    code,
	}
}

var (
	BadRequestHander = func(e error) *Error {
		return buildError(e, http.StatusBadRequest)
	}
	NotFoundHander = func(e error) *Error {
		return buildError(e, http.StatusNotFound)
	}
	InternalServerHandler = func(e error) *Error {
		return buildError(e, http.StatusInternalServerError)
	}
	OKHandler = func() *Error {
		return nil
	}
)
