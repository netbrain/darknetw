package api

import (
	"net/http"
	"sort"
)

type Route map[Methods]http.Handler
type Routes map[string]Route

func (r Routes) Walk(fn func(path string, methods Methods, handler http.Handler) error) error {
	var paths []string
	for path := range r {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		for methods, handler := range r[path] {
			if err := fn(path, methods, handler); err != nil {
				return err
			}
		}
	}
	return nil
}

type Routable interface {
	Routes() Routes
}
