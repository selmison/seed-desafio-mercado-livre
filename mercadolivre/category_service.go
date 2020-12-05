package mercadolivre

import (
	"context"

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
	return Validate(c)
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
	msgError := "service.category_post"
	if err != nil {
		return "", errors.Wrap(err, msgError)
	}
	id := uuid.New().String()
	_, err = stmt.Exec(
		id,
		category.Name,
	)
	if err != nil {
		return "", errors.Wrap(err, msgError)
	}
	return id, nil
}
