package mercadolivre

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// Endpoints collects all of the endpoints.
type Endpoints struct {
	UserPostEndpoint endpoint.Endpoint
}

// MakeServerEndpoints returns an Endpoints struct.
func MakeServerEndpoints(svc Service) Endpoints {
	return Endpoints{
		UserPostEndpoint: ValidationMiddleware()(MakeUserPostEndpoint(svc)),
	}
}

// MakeUserPostEndpoint returns an endpoint via the passed service.
func MakeUserPostEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(postUserRequest)
		id, err := svc.UserPost(ctx, req.User)
		return postResponse{
			Id:  id,
			Err: err,
		}, nil
	}
}

type postUserRequest struct {
	User UserRequest
}

func (p postUserRequest) Validate() error {
	return p.User.Validate()
}

type postResponse struct {
	Id  string
	Err error `json:"err,omitempty"`
}

func (r postResponse) error() error { return r.Err }
