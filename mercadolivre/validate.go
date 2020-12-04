package mercadolivre

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/non-standard/validators"
	"github.com/pkg/errors"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	if err := validate.RegisterValidation("not_blank", validators.NotBlank); err != nil {
		log.Fatalln(err)
	}
	if err := validate.RegisterValidation("should_be_future", ShouldBeFuture); err != nil {
		log.Fatalln(err)
	}
}

type ValidationErrorsResponse []*ValidationErrorResponse

type ValidationErrorResponse struct {
	FailedField string `json:"failed_field"`
	Condition   string `json:"condition"`
	ActualValue string `json:"actual_value"`
}

func (v ValidationErrorsResponse) Error() string {
	return ErrValidationFailed.Error()
}

// ShouldBeFuture validates if the current field is time.Time and is after time.Now().
func ShouldBeFuture(fl validator.FieldLevel) bool {
	field := fl.Field()

	if field.Kind() == reflect.Struct {
		if v, ok := field.Interface().(time.Time); ok {
			now := time.Now()
			if now.Before(v) {
				return true
			}
		}
	}
	return false
}

// ShouldBeUnique validates if the current field value is unique in the repository.
func (s *service) ShouldBeUnique(fl validator.FieldLevel) bool {
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
