package mercadolivre

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/handlers"
)

type httpServer struct {
	logger Logger
}

// NewHTTPServer starts new HTTP server
func NewHTTPServer(cfg Config, svc Service, logger Logger) error {
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	lnAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}

	srv := &httpServer{
		logger: logger,
	}
	router := srv.MakeHTTPHandler(svc)
	loggingHandler := handlers.LoggingHandler(os.Stdout, router)
	fmt.Printf("HTTP server listening on http://%s\n", lnAddr.String())
	if err := http.ListenAndServe(lnAddr.String(), loggingHandler); err != nil {
		return err
	}

	return nil
}
