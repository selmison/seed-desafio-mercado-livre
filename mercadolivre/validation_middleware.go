package mercadolivre

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// ValidationMiddleware valides the requests
func ValidationMiddleware() endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			req := request.(Request)
			if err := req.Validate(); err != nil {
				return nil, err
			}
			return next(ctx, request)
		}
	}
}
