package infra

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/selmison/seed-desafio-mercado-livre/logging"
)

// MethodNotAllowedError should be returned by a handler when the HTTP method is not allowed.
type MethodNotAllowedError struct {
	Method string
	Allow  []string
}

func (e MethodNotAllowedError) Error() string {
	return fmt.Sprintf("method %s not allowed", e.Method)
}

// BadRequestError should be returned by a handler when parameters or the payload are not valid
type BadRequestError struct {
	Reason string
}

func (e BadRequestError) Error() string {
	return fmt.Sprintf("Bad request: %s", e.Reason)
}

// NotFoundError should be returned by a handler when a resource specified does not exist
type NotFoundError struct {
	Reason string
}

func (e NotFoundError) Error() string {
	return e.Reason
}

// CodeWithPayloadError allow returning non HTTP 200
// Error codes while not returning PlainText payload
type CodeWithPayloadError struct {
	Reason      string
	StatusCode  int
	ContentType string
}

func (e CodeWithPayloadError) Error() string {
	return e.Reason
}

type ForbiddenError struct {
}

func (e ForbiddenError) Error() string {
	return "Access is restricted"
}

// HTTPHandlers provides an HTTP api.
type HTTPHandlers struct {
	logger hclog.InterceptLogger
	h      http.Handler
}

// endpoint is a HTTP handler that takes the usual arguments in
// but returns a response object and error, both of which are handled in a
// common manner by the HTTP server.
type endpoint func(resp http.ResponseWriter, req *http.Request) (interface{}, error)

// unboundEndpoint is an endpoint method on a server.
type unboundEndpoint func(s *HTTPHandlers, resp http.ResponseWriter, req *http.Request) (interface{}, error)

// endpoints is a map from URL pattern to unbound endpoint.
var endpoints map[string]unboundEndpoint

// allowedMethods is a map from endpoint prefix to supported HTTP methods.
// An empty slice means an endpoint handles OPTIONS requests and MethodNotFound errors itself.
var allowedMethods map[string][]string = make(map[string][]string)

// registerEndpoint registers a new endpoint, which should be done at package
// init() time.
func registerEndpoint(pattern string, methods []string, fn unboundEndpoint) {
	if endpoints == nil {
		endpoints = make(map[string]unboundEndpoint)
	}
	if endpoints[pattern] != nil || allowedMethods[pattern] != nil {
		panic(fmt.Errorf("Pattern %q is already registered", pattern))
	}

	endpoints[pattern] = fn
	allowedMethods[pattern] = methods
}

// wrappedMux hangs on to the underlying mux for unit tests.
type wrappedMux struct {
	mux     *http.ServeMux
	handler http.Handler
}

// ServeHTTP implements the http.Handler interface.
func (w *wrappedMux) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	w.handler.ServeHTTP(resp, req)
}

// handler is used to initialize the Handler. In agent code we only ever call
// this once during agent initialization so it was always intended as a single
// pass init method. However many test rely on it as a cheaper way to get a
// handler to call ServeHTTP against and end up calling it multiple times on a
// single agent instance. Until this method had to manage state that might be
// affected by a reload or otherwise vary over time that was not problematic
// although it was wasteful to redo all this setup work multiple times in one
// test.
//
// Now uiserver and possibly other components need to handle reloadable state
// having test randomly clobber the state with the original config again for
// each call gets confusing fast. So handler will memoize it's response - it's
// allowed to call it multiple times on the same agent, but it will only do the
// work the first time and return the same handler on subsequent calls.
//
// The `enableDebug` argument used in the first call will be effective and a
// later change will not do anything. The same goes for the initial config. For
// example if config is reloaded with UI enabled but it was not originally, the
// http.Handler returned will still have it disabled.
//
// The first call must not be concurrent with any other call. Subsequent calls
// may be concurrent with HTTP requests since no state is modified.
func (s *HTTPHandlers) handler(enableDebug bool) http.Handler {
	// Memoize multiple calls.
	if s.h != nil {
		return s.h
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", s.Index)

	s.h = mux
	return s.h
}

// wrap is used to wrap functions to make them more convenient
func (s *HTTPHandlers) wrap(handler endpoint, methods []string) http.HandlerFunc {
	httpLogger := s.logger.Named(logging.HTTP)
	return func(resp http.ResponseWriter, req *http.Request) {

		// Obfuscate any tokens from appearing in the logs
		formVals, err := url.ParseQuery(req.URL.RawQuery)
		if err != nil {
			httpLogger.Error("Failed to decode query",
				"from", req.RemoteAddr,
				"error", err,
			)
			resp.WriteHeader(http.StatusInternalServerError)
			return
		}
		logURL := req.URL.String()
		if tokens, ok := formVals["token"]; ok {
			for _, token := range tokens {
				if token == "" {
					logURL += "<hidden>"
					continue
				}
				logURL = strings.Replace(logURL, token, "<hidden>", -1)
			}
		}

		isForbidden := func(err error) bool {
			_, ok := err.(ForbiddenError)
			return ok
		}

		isMethodNotAllowed := func(err error) bool {
			_, ok := err.(MethodNotAllowedError)
			return ok
		}

		isBadRequest := func(err error) bool {
			_, ok := err.(BadRequestError)
			return ok
		}

		isNotFound := func(err error) bool {
			_, ok := err.(NotFoundError)
			return ok
		}

		addAllowHeader := func(methods []string) {
			resp.Header().Add("Allow", strings.Join(methods, ","))
		}

		handleErr := func(err error) {
			httpLogger.Error("Request error",
				"method", req.Method,
				"url", logURL,
				"from", req.RemoteAddr,
				"error", err,
			)
			switch {
			case isForbidden(err):
				resp.WriteHeader(http.StatusForbidden)
				fmt.Fprint(resp, err.Error())
			case isMethodNotAllowed(err):
				// RFC2616 states that for 405 Method Not Allowed the response
				// MUST include an Allow header containing the list of valid
				// methods for the requested resource.
				// https://www.w3.org/Protocols/rfc2616/rfc2616-sec10.html
				addAllowHeader(err.(MethodNotAllowedError).Allow)
				resp.WriteHeader(http.StatusMethodNotAllowed) // 405
				fmt.Fprint(resp, err.Error())
			case isBadRequest(err):
				resp.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(resp, err.Error())
			case isNotFound(err):
				resp.WriteHeader(http.StatusNotFound)
				fmt.Fprint(resp, err.Error())
			default:
				resp.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(resp, err.Error())
			}
		}

		start := time.Now()
		defer func() {
			httpLogger.Debug("Request finished",
				"method", req.Method,
				"url", logURL,
				"from", req.RemoteAddr,
				"latency", time.Since(start).String(),
			)
		}()

		var obj interface{}

		// if this endpoint has declared methods, respond appropriately to OPTIONS requests. Otherwise let the endpoint handle that.
		if req.Method == "OPTIONS" && len(methods) > 0 {
			addAllowHeader(append([]string{"OPTIONS"}, methods...))
			return
		}

		// if this endpoint has declared methods, check the request method. Otherwise let the endpoint handle that.
		methodFound := len(methods) == 0
		for _, method := range methods {
			if method == req.Method {
				methodFound = true
				break
			}
		}

		if !methodFound {
			err = MethodNotAllowedError{req.Method, append([]string{"OPTIONS"}, methods...)}
		} else {
			// Invoke the handler
			obj, err = handler(resp, req)
		}
		contentType := "application/json"
		httpCode := http.StatusOK
		if err != nil {
			if errPayload, ok := err.(CodeWithPayloadError); ok {
				httpCode = errPayload.StatusCode
				if errPayload.ContentType != "" {
					contentType = errPayload.ContentType
				}
				if errPayload.Reason != "" {
					resp.Header().Add("X-Consul-Reason", errPayload.Reason)
				}
			} else {
				handleErr(err)
				return
			}
		}
		if obj == nil {
			return
		}
		var buf []byte
		if contentType == "application/json" {
			buf, err = s.marshalJSON(req, obj)
			if err != nil {
				handleErr(err)
				return
			}
		} else {
			if strings.HasPrefix(contentType, "text/") {
				if val, ok := obj.(string); ok {
					buf = []byte(val)
				}
			}
		}
		resp.Header().Set("Content-Type", contentType)
		resp.WriteHeader(httpCode)
		resp.Write(buf)
	}
}

// marshalJSON marshals the object into JSON, respecting the user's pretty-ness
// configuration.
func (s *HTTPHandlers) marshalJSON(req *http.Request, obj interface{}) ([]byte, error) {
	if _, ok := req.URL.Query()["pretty"]; ok {
		buf, err := json.MarshalIndent(obj, "", "    ")
		if err != nil {
			return nil, err
		}
		buf = append(buf, "\n"...)
		return buf, nil
	}

	buf, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// Renders a simple index page
func (s *HTTPHandlers) Index(resp http.ResponseWriter, req *http.Request) {
	// // Send special headers too since this endpoint isn't wrapped with something
	// // that sends them.
	// setHeaders(resp, s.agent.config.HTTPResponseHeaders)

	// Check if this is a non-index path
	if req.URL.Path != "/" {
		resp.WriteHeader(http.StatusNotFound)
		return
	}
}

// setHeaders is used to set canonical response header fields
func setHeaders(resp http.ResponseWriter, headers map[string]string) {
	for field, value := range headers {
		resp.Header().Set(http.CanonicalHeaderKey(field), value)
	}
}
