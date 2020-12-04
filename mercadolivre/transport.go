package mercadolivre

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-kit/kit/transport"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// MakeHTTPHandler mounts all of the service endpoints into an http.Handler.
// Useful in a usersvc server.
func (srv *httpServer) MakeHTTPHandler(svc Service) http.Handler {
	r := mux.NewRouter()
	e := MakeServerEndpoints(svc)
	errorHandler := func(ctx context.Context, err error) {
		if _, ok := err.(ValidationErrorsResponse); !ok {
			st := srv.retrievStackTrace(err)
			srv.logger.Errorf("%s%+v", err, st)
		}
	}
	options := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(transport.ErrorHandlerFunc(errorHandler)),
		httptransport.ServerErrorEncoder(encodeError),
	}

	r.Methods("POST").Path("/users").Handler(httptransport.NewServer(
		e.UserPostEndpoint,
		decodeUserPostRequest,
		encodePostResponse,
		options...,
	))

	r.Methods("POST").Path("/categories").Handler(httptransport.NewServer(
		e.CategoryPostEndpoint,
		decodeCategoryPostRequest,
		encodePostResponse,
		options...,
	))

	return r
}

func decodeUserPostRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req UserRequest
	if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
		return nil, e
	}
	return req, nil
}

func decodeCategoryPostRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req CategoryRequest
	if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
		return nil, e
	}
	return req, nil
}

// errorer is implemented by all concrete response types that may contain
// errors. It allows us to change the HTTP response code without needing to
// trigger an endpoint (transport-level) error.
type errorer interface {
	error() error
}

func encodePostResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	id := fmt.Sprintf("/%s", response.(postResponse).Id)
	w.Header().Set("Location", id)
	w.WriteHeader(http.StatusCreated)
	return json.NewEncoder(w).Encode(response)
}

func encodeError(ctx context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codeFrom(err))
	if e, ok := err.(ValidationErrorsResponse); ok {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"msg":    e.Error(),
			"errors": e,
		})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": http.StatusText(http.StatusInternalServerError),
	})
}

func codeFrom(err error) int {
	if _, ok := err.(ValidationErrorsResponse); ok {
		return http.StatusBadRequest
	}
	switch err {
	case ErrNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func (srv *httpServer) retrievStackTrace(err error) errors.StackTrace {
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}
	e, ok := err.(stackTracer)
	if !ok {
		srv.logger.Error("err does not implement stackTracer")
	}
	return e.StackTrace()
}
