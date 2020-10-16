package api

import (
	"net/http"
	"time"
)

type Context struct {
	Request  *http.Request
	Response http.ResponseWriter
}

func (c Context) Deadline() (deadline time.Time, ok bool) {
	return c.Request.Context().Deadline()
}

func (c Context) Done() <-chan struct{} {
	return c.Request.Context().Done()
}

func (c Context) Err() error {
	return c.Request.Context().Err()
}

func (c Context) Value(key interface{}) interface{} {
	return c.Request.Context().Value(key)
}
