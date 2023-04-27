//  Copyright Istio Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package http

import (
	"fmt"
	"net/http"
	"strconv"
)

// Response the unify response of the http server
type Response struct {
	Error  *Error      `json:"error,omitempty"`
	Result interface{} `json:"result,omitempty"`
}

// Error error message for the return.
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
	TooManyRequestsHandler = func() *Error {
		return buildError(fmt.Errorf("too many requests"), http.StatusTooManyRequests)
	}
	OKHandler = func() *Error {
		return nil
	}
)
