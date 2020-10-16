package api

import "strings"

type Methods uint16

const (
	GET Methods = 1 << iota
	HEAD
	POST
	PUT
	DELETE
	CONNECT
	OPTIONS
	TRACE
	PATCH
	ALL = GET | HEAD | POST | PUT | DELETE | CONNECT | OPTIONS | TRACE | PATCH
)

func (m Methods) Has(flag Methods) bool {
	return m&flag != 0
}

func (m Methods) String() string {
	switch m {
	case GET:
		return "GET"
	case HEAD:
		return "HEAD"
	case POST:
		return "POST"
	case PUT:
		return "PUT"
	case DELETE:
		return "DELETE"
	case CONNECT:
		return "CONNECT"
	case OPTIONS:
		return "OPTIONS"
	case TRACE:
		return "TRACE"
	case PATCH:
		return "PATCH"
	default:
		return strings.Join(m.StringSlice(), " ")

	}
}

func (m Methods) StringSlice() (methods []string) {
	for _, method := range []Methods{
		GET,
		HEAD,
		POST,
		PUT,
		DELETE,
		CONNECT,
		OPTIONS,
		TRACE,
		PATCH,
	} {
		if m.Has(method) {
			methods = append(methods, method.String())
		}
	}
	return
}
