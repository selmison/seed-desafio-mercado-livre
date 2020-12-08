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
		ReAuthEndpoint:       ValidationMdlwr()(MakeReAuthEndpoint(svc)),
		UserPostEndpoint:     ValidationMdlwr()(MakeUserPostEndpoint(svc)),
	}
}

type postResponse struct {
	Id string
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
			Id: id,
		}, nil
	}
}

// MakeReAuthEndpoint returns an endpoint via the passed service.
func MakeReAuthEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(ReAuthRequest)
		res, err := svc.ReAuth(ctx, req)
		if err != nil {
			return nil, err
		}
		return res, nil
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
			Id: id,
		}, nil
	}
}
