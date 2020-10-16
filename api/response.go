package api

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
)

type Response interface {
	http.Handler
}

type ErrorResponse struct {
	Status     int    `json:"status"`
	StatusText string `json:"statusText,omitempty"`
	ErrorText  string `json:"errorText,omitempty"`
}

type Option func(r *response)

func WithHeader(key, value string) Option {
	return func(r *response) {
		if r.headers == nil {
			r.headers = make(http.Header)
		}
		r.headers.Set(key, value)
	}
}

func WithStatus(status int) Option {
	return func(r *response) {
		r.status = status
	}
}

func WithBody(data []byte) Option {
	return func(r *response) {
		r.body = bytes.NewBuffer(data)
	}
}

type response struct {
	status  int
	headers http.Header
	body    *bytes.Buffer
}

func (r *response) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for k, vs := range r.headers {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	if r.status != 0 {
		w.WriteHeader(r.status)
	}
	if r.body == nil {
		return
	}
	if _, err := r.body.WriteTo(w); err != nil {
		panic(err)
	}
}

func NewResponse(opts ...Option) *response {
	c := &response{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func Status(status int, other ...Option) Response {
	return NewResponse(
		append([]Option{WithStatus(status)}, other...)...,
	)
}

func Error(err error, other ...Option) Response {
	return ErrorString(http.StatusInternalServerError, err.Error(), other...)
}

func ErrorString(code int, err string, other ...Option) Response {
	log.Println(code, err, "\n", string(debug.Stack()))
	return JSON(ErrorResponse{
		Status:     code,
		StatusText: http.StatusText(code),
		ErrorText:  err,
	}, append(other, WithStatus(code))...)
}

func JSON(data interface{}, other ...Option) Response {
	jbuf, err := json.Marshal(data)
	if err != nil {
		return Error(err, other...)
	}
	return JSONRaw(jbuf, other...)
}

func JSONRaw(data []byte, other ...Option) Response {
	return NewResponse(
		append([]Option{
			WithHeader("Content-Type", "application/json"),
			WithBody(data),
		}, other...)...,
	)
}

func BadRequest() Response {
	return Status(http.StatusBadRequest)
}

func NotFound() Response {
	return Status(http.StatusNotFound)
}

func Redirect(url string, status int) Response {
	return NewResponse(
		WithStatus(status),
		WithHeader("Location", url),
	)
}

func OK() Response {
	return nil
}
