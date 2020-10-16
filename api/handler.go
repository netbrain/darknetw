package api

import (
	"fmt"
	"log"
	"net/http"
)

type HandlerFn func(ctx Context) Response

func (c HandlerFn) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		rec := recover()
		if rec != nil {
			if err, ok := rec.(error); ok {
				err = fmt.Errorf("panic caught in handler middleware: %w", err)
				log.Println(err)
				Error(err).ServeHTTP(w, r)
				return
			}
			e := fmt.Sprintf("panic caught in controller middleware: %s", rec)
			log.Println(e)
			ErrorString(http.StatusInternalServerError, e).ServeHTTP(w, r)
		}
	}()
	resp := c(Context{
		Request:  r,
		Response: w,
	})
	if resp == nil {
		return
	}
	resp.ServeHTTP(w, r)
}
