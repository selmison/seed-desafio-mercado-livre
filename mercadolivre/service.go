package mercadolivre

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

type Request interface {
	Validate() error
}

type UserRequest struct {
	Name     string `validate:"required,not_blank,email"`
	Password string `validate:"required,not_blank"`
}

type UserResponse struct {
	Name      string
	CreatedAt time.Time
}

func (u *UserRequest) Validate() error {
	err := validate.Struct(u)
	var errs ValidationErrorsResponse
	if err != nil {
		if fieldError, ok := err.(validator.ValidationErrors); ok {
			for _, v := range fieldError {
				var element ValidationErrorResponse
				element.FailedField = v.StructNamespace()
				element.Condition = v.Tag()
				element.ActualValue = v.Value().(string)
				errs = append(errs, &element)
			}
		}
	}
	if err != nil {
		return errs
	}
	return nil
}

// Service is a simple CRUD interface for user.
type Service interface {
	UserPost(ctx context.Context, u UserRequest) (id string, err error)
}

// User represents a single user.
// ID should be globally unique.
type User struct {
	Name      string
	password  string
	CreatedAt time.Time
}

type service struct {
	db     *sqlx.DB
	logger Logger
}

func NewService(driverName, dsn string, logger Logger) (Service, error) {
	db, err := sqlx.Connect(driverName, dsn)
	if err != nil {
		return nil, err
	}
	return &service{
		db:     db,
		logger: logger,
	}, nil
}

func (s *service) UserPost(ctx context.Context, user UserRequest) (string, error) {
	insertUser := "INSERT ITO users (id, name, password, created_at) VALUES ($1, $2, $3, $4)"
	now := time.Now()
	layout := "2006-01-02 15:04:05"
	id := uuid.New().String()
	_, err := s.db.Exec(
		insertUser,
		id,
		user.Name,
		user.Password,
		now.Format(layout))
	if err != nil {
		return "", errors.Wrap(err, "service.post_user")
	}
	return id, nil
}
