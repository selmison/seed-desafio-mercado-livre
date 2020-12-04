package mercadolivre

import (
	"context"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type CategoryRequest struct {
	Name string `validate:"required,not_blank,should_be_unique"`
}

type CategoryResponse struct {
	ID   string
	Name string
}

// Validate validates CategoryRequest.
func (c CategoryRequest) Validate() error {
	err := validate.Struct(c)
	var errs ValidationErrorsResponse
	if err != nil {
		if fieldError, ok := err.(validator.ValidationErrors); ok {
			for _, v := range fieldError {
				element := ValidationErrorResponse{
					FailedField: strings.ToLower(v.StructNamespace()),
					Condition:   v.Tag(),
					ActualValue: v.Value().(string),
				}
				errs = append(errs, &element)
			}
		}
	}
	if err != nil {
		return errs
	}
	return nil
}

// Category represents a single Category.
// ID should be globally unique.
type Category struct {
	ID   string
	Name string
}

// CategoryPost creates category.
func (s *service) CategoryPost(ctx context.Context, category CategoryRequest) (string, error) {
	stmt, err := s.db.Prepare("INSERT INTO categories (id, name) VALUES ($1, $2)")
	if err != nil {
		return "", errors.Wrap(err, "service.post_category")
	}
	id := uuid.New().String()
	_, err = stmt.Exec(
		id,
		category.Name,
	)
	if err != nil {
		return "", errors.Wrap(err, "service.post_category")
	}
	return id, nil
}
