package mercadolivre

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// ValidationMdlwr valides the requests.
func ValidationMdlwr() endpoint.Middleware {
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
