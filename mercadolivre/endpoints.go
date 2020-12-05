package mercadolivre

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// Endpoints collects all of the endpoints.
type Endpoints struct {
	AuthEndpoint         endpoint.Endpoint
	CategoryPostEndpoint endpoint.Endpoint
	UserPostEndpoint     endpoint.Endpoint
}

// MakeServerEndpoints returns an Endpoints struct.
func MakeServerEndpoints(svc Service) Endpoints {
	return Endpoints{
		AuthEndpoint:         ValidationMiddleware()(MakeAuthEndpoint(svc)),
		CategoryPostEndpoint: ValidationMiddleware()(MakeCategoryPostEndpoint(svc)),
		UserPostEndpoint:     ValidationMiddleware()(MakeUserPostEndpoint(svc)),
	}
}

type postResponse struct {
	Id string
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
