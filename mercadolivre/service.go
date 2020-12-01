package mercadolivre

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/multierr"
)

type Request interface {
	Validate() error
}

type UserRequest struct {
	Name     string `validate:"required,email"`
	Password string `validate:"required"`
}

type UserResponse struct {
	Name      string
	CreatedAt time.Time
}

func (u *UserRequest) Validate() error {
	err := validate.Struct(u)
	if err != nil {
		if vErrs, ok := err.(validator.ValidationErrors); ok {
			var err error
			for _, v := range vErrs {
				err = multierr.Append(
					err,
					fmt.Errorf("the '%s' field %w", v.Namespace(), ErrIsNotValid),
				)
			}
		}
	}
	if err != nil {
		return err
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

func (s *service) UserPost(ctx context.Context, user UserRequest) (id string, err error) {
	insertUser := "INSERT INTO users (id, name, password, created_at) VALUES ($1, $2, $3, $4)"
	now := time.Now()
	layout := "2006-01-02 15:04:05"
	id = uuid.New().String()
	_, err = s.db.Exec(
		insertUser,
		id,
		user.Name,
		user.Password,
		now.Format(layout))
	if err != nil {
		s.logger.Errorf("service.post_user: %v\n", err)
		return
	}
	return
}
