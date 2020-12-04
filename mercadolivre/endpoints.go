package mercadolivre

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// Endpoints collects all of the endpoints.
type Endpoints struct {
	UserPostEndpoint     endpoint.Endpoint
	CategoryPostEndpoint endpoint.Endpoint
}

// MakeServerEndpoints returns an Endpoints struct.
func MakeServerEndpoints(svc Service) Endpoints {
	return Endpoints{
		UserPostEndpoint:     ValidationMiddleware()(MakeUserPostEndpoint(svc)),
		CategoryPostEndpoint: ValidationMiddleware()(MakeCategoryPostEndpoint(svc)),
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

type postResponse struct {
	Id  string
	Err error `json:"err,omitempty"`
}

func (r postResponse) error() error { return r.Err }

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
