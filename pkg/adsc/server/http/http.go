package http

import (
	"encoding/json"
	"math"
	"net/http"
)

const abortIndex int = math.MaxInt / 2

type Server struct {
	address string
	router  *router
}

func New(address string) *Server {
	return &Server{
		address: address,
		router: &router{
			mux: http.NewServeMux(),
		},
	}
}

func (s *Server) Use(middleware ...HandlerFunc) *Server {
	s.router.middleware = append(s.router.middleware, middleware...)
	return s
}

func (s *Server) Register(pattern string, h Handler) *Server {
	s.router.mux.Handle(pattern, buildHandler(h))
	return s
}

func (s *Server) Serve() error {
	httpServer := &http.Server{
		Handler: s.router,
		Addr:    s.address,
	}
	return httpServer.ListenAndServe()
}

type router struct {
	middleware []HandlerFunc
	mux        *http.ServeMux
}

type HandleContext struct {
	index      int
	middleware []HandlerFunc
	mux        *http.ServeMux
	Writer     ResponseWriter
	Request    *http.Request
}

func (c *HandleContext) Abort() {
	c.index = abortIndex
}

func (c *HandleContext) Next() {
	c.index++
	for c.index <= len(c.middleware) {
		if c.index == len(c.middleware) {
			c.mux.ServeHTTP(c.Writer, c.Request)
			c.index++
			break
		}
		c.middleware[c.index](c)
		c.index++
	}
}

type HandlerFunc func(*HandleContext)
type Handler func(request *http.Request) (interface{}, *Error)

func (r *router) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	c := &HandleContext{
		middleware: r.middleware,
		mux:        r.mux,
		Writer: &mcpResponse{
			w: w,
		},
		Request: request,
		index:   -1,
	}
	c.Next()
}

func handleResponse(w http.ResponseWriter, ret interface{}, err *Error) {
	resp := &Response{
		Error:  err,
		Result: ret,
	}
	out, _ := json.Marshal(resp)
	w.WriteHeader(err.GetCode())
	w.Write(out)
}

func buildHandler(fn Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		r, err := fn(request)
		handleResponse(w, r, err)
	})
}

// ResponseWriter the wrapper of http ResponseWriter
type ResponseWriter interface {
	StatusCode() int
	Done()
	WriteError(*Error)
	http.ResponseWriter
}

type mcpResponse struct {
	w        http.ResponseWriter
	code     int
	response []byte
}

func (mr *mcpResponse) WriteError(err *Error) {
	resp := &Response{
		Error: err,
	}
	mr.response, _ = json.Marshal(resp)
	mr.code = err.GetCode()
}

func (mr *mcpResponse) Done() {
	mr.w.WriteHeader(mr.code)
	mr.w.Write(mr.response)
}

func (mr *mcpResponse) Header() http.Header {
	return mr.w.Header()
}

func (mr *mcpResponse) Write(bs []byte) (int, error) {
	mr.response = bs
	return len(bs), nil
}
func (mr *mcpResponse) WriteHeader(statusCode int) {
	mr.code = statusCode
}

func (mr *mcpResponse) StatusCode() int {
	return mr.code
}
