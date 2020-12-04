package mercadolivre

import (
	"context"
	"database/sql"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Request interface {
	Validate() error
}

// Service is a simple CRUD interface for user.
type Service interface {
	UserPost(ctx context.Context, u UserRequest) (id string, err error)
	CategoryPost(ctx context.Context, u CategoryRequest) (id string, err error)
}

type service struct {
	validate *validator.Validate
	db       *sqlx.DB
	logger   Logger
}

// NewService creates a service with the necessary dependencies.
func NewService(db *sql.DB, driverName string, logger Logger) (Service, error) {
	dbx := sqlx.NewDb(db, driverName)
	if err := dbx.Ping(); err != nil {
		return nil, err
	}

	svc := &service{
		validate: validate,
		db:       dbx,
		logger:   logger,
	}

	if err := validate.RegisterValidation("should_be_unique", svc.ShouldBeUnique); err != nil {
		logger.Fatal(err)
	}

	return svc, nil
}
