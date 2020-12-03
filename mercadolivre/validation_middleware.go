package mercadolivre

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// ValidationMiddleware valides the requests
func (srv *httpServer) ValidationMiddleware() endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			req := request.(Request)
			if err := req.Validate(); err != nil {
				return nil, err
			}
			return next(ctx, request)
		}
	}
}
