package infra

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/selmison/seed-desafio-mercado-livre/cmd"
)

// Server is the http server
type Server struct {
	// Consul configuration
	config *cmd.Config
	// Logger uses the provided LogOutput
	logger hclog.InterceptLogger

	// Listener is used to listen for incoming connections
	Listener net.Listener
}

func Start(ctx context.Context) (*Server, error) {
	cfg, err := cmd.NewConfig()
	if err != nil {
		return nil, err
	}
	server, err := NewServer(cfg, a.baseDeps.Deps)
	if err != nil {
		return nil, fmt.Errorf("Failed to start Consul server: %v", err)
	}
	return server, nil
}

// NewServer is used to construct a new server from the configuration
func NewServer(config *cmd.Config, logger hclog.InterceptLogger) (*Server, error) {
	s := &Server{
		config: config,
		logger: logger,
	}

	srv := &HTTPHandlers{
		logger: logger,
	}
	a.configReloaders = append(a.configReloaders, srv.ReloadConfig)
	a.httpHandlers = srv
	httpServer := &http.Server{
		Addr:      l.Addr().String(),
		TLSConfig: tlscfg,
		Handler:   srv.handler(a.config.EnableDebug),
	}

	go s.listen(s.Listener)

	return s, nil
}

// listenHTTP binds listeners to the provided addresses and also returns
// pre-configured HTTP servers which are not yet started. The motivation is
// that in the current startup/shutdown setup we de-couple the listener
// creation from the server startup assuming that if any of the listeners
// cannot be bound we fail immediately and later failures do not occur.
// Therefore, starting a server with a running listener is assumed to not
// produce an error.
//
// The second motivation is that an HTTPS server needs to use the same TLSConfig
// on both the listener and the HTTP server. When listeners and servers are
// created at different times this becomes difficult to handle without keeping
// the TLS configuration somewhere or recreating it.
//
// This approach should ultimately be refactored to the point where we just
// start the server and any error should trigger a proper shutdown of the agent.
func (s *Server) listenHTTP() error {
	var ln []net.Listener

	start := func(proto string, addrs []net.Addr) error {
		listeners, err := s.startListeners(addrs)
		if err != nil {
			return err
		}
		ln = append(ln, listeners...)

		for _, l := range listeners {
			var tlscfg *tls.Config
			_, isTCP := l.(*tcpKeepAliveListener)
			if isTCP && proto == "https" {
				tlscfg = s.tlsConfigurator.IncomingHTTPSConfig()
				l = tls.NewListener(l, tlscfg)
			}

			srv := &HTTPHandlers{
				agent:    s,
				denylist: NewDenylist(s.config.HTTPBlockEndpoints),
			}
			s.configReloaders = append(s.configReloaders, srv.ReloadConfig)
			s.httpHandlers = srv
			httpServer := &http.Server{
				Addr:      l.Addr().String(),
				TLSConfig: tlscfg,
				Handler:   srv.handler(s.config.EnableDebug),
			}

			// Load the connlimit helper into the server
			connLimitFn := s.httpConnLimiter.HTTPConnStateFuncWithDefault429Handler(10 * time.Millisecond)

			if proto == "https" {
				if err := setupHTTPS(httpServer, connLimitFn, s.config.HTTPSHandshakeTimeout); err != nil {
					return err
				}
			} else {
				httpServer.ConnState = connLimitFn
			}

			servers = append(servers, newAPIServerHTTP(proto, l, httpServer))
		}
		return nil
	}

	if err := start("http", s.config.HTTPAddrs); err != nil {
		closeListeners(ln)
		return nil, err
	}
	if err := start("https", s.config.HTTPSAddrs); err != nil {
		closeListeners(ln)
		return nil, err
	}
	return servers, nil
}
