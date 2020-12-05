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
			if errors.Is(err, ErrLoginFailed) {
				srv.logger.Warn(err)
			} else {
				st := srv.retrieveStackTrace(err)
				srv.logger.Errorf("%s%+v", err, st)
			}
		}
	}
	options := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(transport.ErrorHandlerFunc(errorHandler)),
		httptransport.ServerErrorEncoder(encodeError),
	}

	r.Methods("POST").Path("/categories").Handler(httptransport.NewServer(
		e.CategoryPostEndpoint,
		decodeCategoryPostRequest,
		encodePostResponse,
		options...,
	))

	r.Methods("POST").Path("/login").Handler(httptransport.NewServer(
		e.LoginPostEndpoint,
		decodeLoginPostRequest,
		encodeLoginResponse,
		options...,
	))

	r.Methods("POST").Path("/users").Handler(httptransport.NewServer(
		e.UserPostEndpoint,
		decodeUserPostRequest,
		encodePostResponse,
		options...,
	))

	return r
}

func decodeCategoryPostRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req CategoryRequest
	if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
		return nil, e
	}
	return req, nil
}

func decodeLoginPostRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req LoginRequest
	if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
		return nil, e
	}
	return req, nil
}

func decodeUserPostRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req UserRequest
	if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
		return nil, e
	}
	return req, nil
}

func encodeLoginResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	login := response.(LoginResponse)
	http.SetCookie(w,
		&http.Cookie{
			Name:    "token",
			Value:   login.Token,
			Expires: login.ExpiresAt,
		})
	return nil
}

func encodePostResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
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
	statusCode := codeFrom(err)
	w.WriteHeader(statusCode)
	if e, ok := err.(ValidationErrorsResponse); ok {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"msg":    e.Error(),
			"errors": e,
		})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": http.StatusText(statusCode),
	})
}

func codeFrom(err error) int {
	if _, ok := err.(ValidationErrorsResponse); ok {
		return http.StatusBadRequest
	}
	if errors.Is(err, ErrNotFound) {
		return http.StatusNotFound
	}
	if errors.Is(err, ErrLoginFailed) {
		return http.StatusUnauthorized
	}

	return http.StatusInternalServerError
}

func (srv *httpServer) retrieveStackTrace(err error) errors.StackTrace {
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}
	e, ok := err.(stackTracer)
	if !ok {
		srv.logger.Error("err does not implement stackTracer")
	}
	return e.StackTrace()
}
