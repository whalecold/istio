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

package server

import (
	"fmt"
	"strconv"
	"time"

	adscmetrics "istio.io/istio/pkg/adsc/metrics"
	http2 "istio.io/istio/pkg/adsc/server/http"
)

func (s *server) Limiter() func(*http2.HandleContext) {
	return func(ctx *http2.HandleContext) {
		if !s.limiter.Allow() {
			ctx.Abort()
			ctx.Writer.WriteError(http2.TooManyRequestsHandler())
			return
		}
		ctx.Next()
	}
}

func responser(ctx *http2.HandleContext) {
	ctx.Next()
	ctx.Writer.Done()
}

func recovery(ctx *http2.HandleContext) {
	defer func() {
		if e := recover(); e != nil {
			ctx.Writer.WriteError(http2.InternalServerHandler(fmt.Errorf("%v", e)))
		}
	}()
	ctx.Next()
}

func observer(ctx *http2.HandleContext) {
	t1 := time.Now()

	ctx.Next()

	code := strconv.Itoa(ctx.Writer.StatusCode())
	t := time.Since(t1).Seconds()
	path := ctx.Request.URL.Path
	kind := stringDef(ctx.Request.URL.Query().Get(queryParameterKind), serviceEntryKind)

	adscmetrics.MCPServerRequestsDuration.WithLabelValues(kind, path, code).Observe(t)
	adscmetrics.MCPServerRequestsTotal.WithLabelValues(kind, path, code).Inc()

	mcplog.Debugf("get resource, kind %v, code: %s req path: %s, query: %s, latency %v", kind, code, path, ctx.Request.URL.RawQuery, t)
}
