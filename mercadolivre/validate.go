package mercadolivre

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
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
func (s service) ShouldBeUnique(fl validator.FieldLevel) bool {
	field := fl.Field()
	fieldName := fl.FieldName()
	var fieldValue string
	if field.Kind() == reflect.String {
		fieldValue = field.String()
	} else {
		return false
	}

	query := fmt.Sprintf(`SELECT %s FROM users WHERE %s=$1`, fieldName, fieldName)
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return false
	}
	v := field.Interface()
	err = stmt.QueryRow(fieldValue).Scan(&v)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return true
		}
	}
	return false
}
