package mercadolivre

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"

	"github.com/go-kit/kit/endpoint"
	"github.com/pkg/errors"
)

// ValidationMiddleware valides the requests
func (srv *httpServer) ValidationMiddleware() endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			req := request.(Request)
			if err := req.Validate(); err != nil {
				return nil, err
			}
			switch v := request.(type) {
			case postUserRequest:
				if err := srv.fieldShouldBeUnique("name", v.User.Name, ""); err != nil {
					return nil, err
				}
			}
			return next(ctx, request)
		}
	}
}

func (srv *httpServer) fieldShouldBeUnique(fieldName, fieldValue string, iface interface{}) error {
	query := fmt.Sprintf(`SELECT %s FROM users WHERE %s=$1`, fieldName, fieldName)
	stmt, err := srv.db.Prepare(query)
	if err != nil {
		return errors.Wrap(err, "http_server.email_validation")
	}
	v := reflect.ValueOf(iface).Interface()
	err = stmt.QueryRow(fieldValue).Scan(&v)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return errors.Wrap(err, "http_server.email_validation")
	}
	err = ValidationErrorsResponse{
		&ValidationErrorResponse{
			FailedField: fieldName,
			Condition:   ErrShouldBeUnique.Error(),
			ActualValue: fieldValue,
		},
	}
	return err
}
