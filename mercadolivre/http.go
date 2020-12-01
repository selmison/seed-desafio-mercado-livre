package mercadolivre

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/handlers"
)

const (
	// ErrInvalidMethod is used if the HTTP method is not supported
	ErrInvalidMethod = "Invalid method"
)

// HTTPServer is used to wrap an Service and expose it over an HTTP interface
// type HTTPServer struct {
// 	logger  Logger
// 	Addr    string
// }

// NewHTTPServer starts new HTTP server
func NewHTTPServer(svc Service, logger Logger, config *Config) error {
	addr := net.JoinHostPort(config.Host, strconv.Itoa(config.Port))
	lnAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}
	// srv := &HTTPServer{
	// 	Addr:    lnAddr.String(),
	// }
	// srv.registerHandlers()

	router := MakeHTTPHandler(svc, logger)
	loggingHandler := handlers.LoggingHandler(os.Stdout, router)
	fmt.Printf("HTTP server listening on http://%s\n", lnAddr.String())
	if err := http.ListenAndServe(lnAddr.String(), loggingHandler); err != nil {
		return err
	}

	return nil
}

// // registerHandlers is used to attach our handlers to the router
// func (s *HTTPServer) registerHandlers() {
// 	s.router.Handle("/", s.handleRoot())
// }

// func (s *HTTPServer) handleRoot() http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
// 		if req.URL.Path == "/" {
// 			if _, err := w.Write([]byte("Welcome!")); err != nil {
// 				s.logger.Errorf("root handler: #v\n", err)
// 			}
// 		} else {
// 			w.WriteHeader(http.StatusNotFound)
// 		}
// 	})
// }
