package mercadolivre

import (
	"context"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type UserRequest struct {
	Name     string `validate:"required,not_blank,email,should_be_unique"`
	Password string `validate:"required,not_blank,min=6"`
}

type UserResponse struct {
	Name      string
	CreatedAt time.Time
}

// Validate validates UserRequest.
func (u UserRequest) Validate() error {
	err := validate.Struct(u)
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

// User represents a single user.
// ID should be globally unique.
type User struct {
	Name      string
	Password  string
	CreatedAt time.Time
}

// UserPost creates user.
func (s *service) UserPost(ctx context.Context, user UserRequest) (string, error) {
	stmt, err := s.db.Prepare("INSERT INTO users (id, name, password, created_at) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return "", errors.Wrap(err, "service.post_user")
	}
	now := time.Now()
	layout := "2006-01-02 15:04:05"
	id := uuid.New().String()
	var hash string
	hash, err = hashPassword(user.Password)
	if err != nil {
		return "", errors.Wrap(err, "service.post_user")
	}
	_, err = stmt.Exec(
		id,
		user.Name,
		hash,
		now.Format(layout))
	if err != nil {
		return "", errors.Wrap(err, "service.post_user")
	}
	return id, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
