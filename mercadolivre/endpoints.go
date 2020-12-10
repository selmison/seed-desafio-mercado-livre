package mercadolivre

import (
	"context"

	"github.com/dgrijalva/jwt-go"
	jwtKit "github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/endpoint"
)

// Endpoints collects all of the endpoints.
type Endpoints struct {
	AuthEndpoint         endpoint.Endpoint
	CategoryPostEndpoint endpoint.Endpoint
	ProductPostEndpoint  endpoint.Endpoint
	ReAuthEndpoint       endpoint.Endpoint
	UserPostEndpoint     endpoint.Endpoint
}

// MakeServerEndpoints returns an Endpoints struct.
func MakeServerEndpoints(svc Service) Endpoints {
	kf := func(token *jwt.Token) (interface{}, error) { return []byte("myJWTSecretKey"), nil }
	AuthMdlwr := jwtKit.NewParser(kf, jwt.SigningMethodHS256, jwtKit.StandardClaimsFactory)

	return Endpoints{
		AuthEndpoint:         ValidationMdlwr()(MakeAuthEndpoint(svc)),
		CategoryPostEndpoint: AuthMdlwr(ValidationMdlwr()(MakeCategoryPostEndpoint(svc))),
		ProductPostEndpoint:  AuthMdlwr(ValidationMdlwr()(MakeProductPostEndpoint(svc))),
		ReAuthEndpoint:       (MakeReAuthEndpoint(svc)),
		UserPostEndpoint:     ValidationMdlwr()(MakeUserPostEndpoint(svc)),
	}
}

type postResponse struct {
	ID string `json:"id"`
}

// MakeAuthEndpoint returns an endpoint via the passed service.
func MakeAuthEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(AuthRequest)
		res, err := svc.Auth(ctx, req)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}

// MakeCategoryPostEndpoint returns an endpoint via the passed service.
func MakeCategoryPostEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(CategoryRequest)
		id, err := svc.CategoryPost(ctx, req)
		if err != nil {
			return nil, err
		}
		return postResponse{
			ID: id,
		}, nil
	}
}

// MakeReAuthEndpoint returns an endpoint via the passed service.
func MakeReAuthEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		res, err := svc.ReAuth(ctx)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}

// MakeProductPostEndpoint returns an endpoint via the passed service.
func MakeProductPostEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(ProductRequest)
		id, err := svc.ProductPost(ctx, req)
		if err != nil {
			return nil, err
		}
		return postResponse{
			ID: id,
		}, nil
	}
}

// MakeUserPostEndpoint returns an endpoint via the passed service.
func MakeUserPostEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(UserRequest)
		id, err := svc.UserPost(ctx, req)
		if err != nil {
			return nil, err
		}
		return postResponse{
			ID: id,
		}, nil
	}
}
