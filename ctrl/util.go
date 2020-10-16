package ctrl

import (
	"bufio"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/netbrain/darknetw/api"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
)

func ReadMultipart(r *http.Request) (*multipart.Reader, error) {
	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(mediaType, "multipart") {
		return nil, fmt.Errorf("not a multipart request")
	}

	boundary, ok := params["boundary"]
	if !ok {
		return nil, fmt.Errorf("no boundary defined")
	}

	return multipart.NewReader(bufio.NewReader(r.Body), boundary), nil
}

func tryToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Println(err)
		return 0
	}
	return i
}

func tryToFloat(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Println(err)
		return 0
	}
	return v
}

func CreateRouter(routable ...api.Routable) http.HandlerFunc {
	router := mux.NewRouter()
	for _, r := range routable {
		r.Routes().Walk(func(path string, methods api.Methods, handler http.Handler) error {
			log.Printf("registering route: [%s] %s", methods, path)
			router.Handle(path, handler).Methods(methods.StringSlice()...)
			return nil
		})
	}
	return router.ServeHTTP
}

func Do(handler http.HandlerFunc, r *http.Request) *http.Response {
	w := httptest.NewRecorder()
	/*rbuf := &bytes.Buffer{}
	r.Clone(context.Background()).Write(rbuf)
	if rbuf.Len() > 512{
		rbuf.Truncate(512)
	}
	fmt.Println(rbuf.String())
	fmt.Println()*/
	handler(w, r)
	/*wbuf := &bytes.Buffer{}
	w.Result().Write(wbuf)
	if wbuf.Len() > 512{
		wbuf.Truncate(512)
	}
	fmt.Println(wbuf.String())
	fmt.Println()*/
	return w.Result()
}
