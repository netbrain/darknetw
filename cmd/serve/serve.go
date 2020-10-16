package serve

import (
	"github.com/netbrain/darknetw/api"
	"github.com/netbrain/darknetw/cfg"
	"github.com/netbrain/darknetw/ctrl"
	"net/http"
	"strings"
	"time"
)

func Run(config *cfg.AppConfig) error {
	router := ctrl.CreateRouter([]api.Routable{
		ctrl.NewDarknetController(config),
	}...)

	srv := &http.Server{
		//Handler: handlers.LoggingHandler(os.Stdout, router),
		Handler:      router,
		Addr:         strings.Join(append([]string{}, config.Host, config.Port), ":"),
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 0,
	}

	return srv.ListenAndServe()
}
