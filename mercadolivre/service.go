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
	Auth(ctx context.Context, req AuthRequest) (*AuthResponse, error)
	CategoryPost(ctx context.Context, req CategoryRequest) (id string, err error)
	ProductPost(ctx context.Context, req ProductRequest) (id string, err error)
	ReAuth(ctx context.Context) (*AuthResponse, error)
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
	if err := validate.RegisterValidation("should_exist", svc.shouldExist); err != nil {
		logger.Fatal(err)
	}

	return svc, nil
}

// shouldExist validates if the current field value exists in the repository.
func (s *service) shouldExist(fl validator.FieldLevel) bool {
	field := fl.Field()
	var fieldValue string
	if field.Kind() == reflect.String {
		fieldValue = field.String()
	} else {
		return false
	}

	fieldName := fl.FieldName()
	var table string
	if fieldName == "CategoryID" {
		table = "categories"
		fieldName = "id"
	}

	query := fmt.Sprintf(`SELECT * FROM %s WHERE %s=$1`, table, fieldName)
	stmt, err := s.db.Preparex(query)
	if err != nil {
		s.logger.Fatal("should_exist validate:", err)
	}

	m := make(map[string]interface{})
	err = stmt.QueryRowx(fieldValue).MapScan(m)
	return err == nil
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
		s.logger.Fatal("should_be_unique validate:", err)
	}

	m := make(map[string]interface{})
	err = stmt.QueryRowx(fieldValue).MapScan(m)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return true
		}
		s.logger.Fatal("should_be_unique validate:", err)
	}
	return false
}
