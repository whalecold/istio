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
