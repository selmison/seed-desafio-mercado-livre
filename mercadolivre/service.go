package mercadolivre

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

type Request interface {
	Validate() error
}

// Service is a simple CRUD interface for user.
type Service interface {
	CategoryPost(ctx context.Context, req CategoryRequest) (id string, err error)
	Auth(ctx context.Context, req AuthRequest) (*AuthResponse, error)
	UserPost(ctx context.Context, req UserRequest) (id string, err error)
}

type service struct {
	validate *validator.Validate
	db       *sqlx.DB
	logger   Logger
}

// NewService creates a service with the necessary dependencies.
func NewService(cfg Config, logger Logger) (Service, error) {
	dbx := sqlx.NewDb(cfg.DB, cfg.DriverName)
	if err := dbx.Ping(); err != nil {
		return nil, err
	}

	svc := &service{
		validate: validate,
		db:       dbx,
		logger:   logger,
	}

	if err := validate.RegisterValidation("should_be_unique", svc.shouldBeUnique); err != nil {
		logger.Fatal(err)
	}

	return svc, nil
}

// shouldBeUnique validates if the current field value is unique in the repository.
func (s *service) shouldBeUnique(fl validator.FieldLevel) bool {
	field := fl.Field()
	fieldName := strings.ToLower(fl.FieldName())
	var fieldValue string
	if field.Kind() == reflect.String {
		fieldValue = field.String()
	} else {
		return false
	}

	var table string
	switch fl.Top().Type().Name() {
	case "CategoryRequest":
		table = "categories"
	case "UserRequest":
		table = "users"
	}
	query := fmt.Sprintf(`SELECT * FROM %s WHERE %s=$1`, table, fieldName)
	stmt, err := s.db.Preparex(query)
	if err != nil {
		return false
	}

	m := make(map[string]interface{})
	err = stmt.QueryRowx(fieldValue).MapScan(m)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return true
		}
	}
	return false
}
