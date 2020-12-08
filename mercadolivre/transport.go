package mercadolivre

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	jwtKit "github.com/go-kit/kit/auth/jwt"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// MakeHTTPHandler mounts all of the service endpoints into an http.Handler.
// Useful in a usersvc server.
func (srv *httpServer) MakeHTTPHandler(svc Service) http.Handler {
	r := mux.NewRouter()
	e := MakeServerEndpoints(svc)

	var jwtTokenRequestFunc httptransport.RequestFunc = func(ctx context.Context, r *http.Request) context.Context {
		c, err := r.Cookie("token")
		if err != nil {
			return ctx
		}

		tknStr := c.Value
		return context.WithValue(ctx, jwtKit.JWTTokenContextKey, tknStr)
	}

	options := []httptransport.ServerOption{
		httptransport.ServerBefore(jwtTokenRequestFunc),
		httptransport.ServerErrorEncoder(srv.encodeError),
	}

	r.Methods("POST").Path("/auth").Handler(httptransport.NewServer(
		e.AuthEndpoint,
		decodeAuthPostRequest,
		encodeAuthResponse,
		options...,
	))

	r.Methods("POST").Path("/categories").Handler(httptransport.NewServer(
		e.CategoryPostEndpoint,
		decodeCategoryPostRequest,
		encodePostResponse,
		options...,
	))

	r.Methods("POST").Path("/reauth").Handler(httptransport.NewServer(
		e.ReAuthEndpoint,
		decodeReAuthPostRequest,
		encodeReAuthResponse,
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

func decodeAuthPostRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	r = r.WithContext(ctx)
	var req AuthRequest
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

func decodeReAuthPostRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req ReAuthRequest
	if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
		return nil, e
	}
	return req, nil
}

func encodeAuthResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	auth := response.(*AuthResponse)
	http.SetCookie(w,
		&http.Cookie{
			Name:    "token",
			Value:   auth.TknStr,
			Expires: auth.ExpiresAt,
		})
	return nil
}

func decodeUserPostRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req UserRequest
	if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
		return nil, e
	}
	return req, nil
}

func encodeReAuthResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	auth := response.(*AuthResponse)
	http.SetCookie(w,
		&http.Cookie{
			Name:    "token",
			Value:   auth.TknStr,
			Expires: auth.ExpiresAt,
		})
	return nil
}

func encodePostResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	id := fmt.Sprintf("/%s", response.(postResponse).Id)
	w.Header().Set("Location", id)
	w.WriteHeader(http.StatusCreated)
	return json.NewEncoder(w).Encode(response)
}

func (srv *httpServer) encodeError(ctx context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	statusCode := srv.logAndCodeFrom(err)
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

func (srv *httpServer) logAndCodeFrom(err error) int {
	if _, ok := err.(ValidationErrorsResponse); ok {
		return http.StatusBadRequest
	}
	if errors.Is(err, ErrAuthFailed) ||
		errors.Is(err, jwtKit.ErrTokenContextMissing) ||
		errors.Is(err, jwtKit.ErrUnexpectedSigningMethod) ||
		errors.Is(err, jwtKit.ErrTokenMalformed) ||
		errors.Is(err, jwtKit.ErrTokenExpired) ||
		errors.Is(err, jwtKit.ErrTokenNotActive) {
		srv.logger.Warn(err)
		return http.StatusUnauthorized
	}

	srv.logStackTrace(err)

	if errors.Is(err, ErrNotFound) {
		return http.StatusNotFound
	}

	return http.StatusInternalServerError
}

func (srv *httpServer) logStackTrace(err error) {
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}
	if e, ok := err.(stackTracer); ok {
		srv.logger.Errorf("%s%+v", err, e.StackTrace())
		return
	}
	srv.logger.Error(err)
}
