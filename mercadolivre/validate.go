package mercadolivre

import (
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/non-standard/validators"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	if err := validate.RegisterValidation("not_blank", validators.NotBlank); err != nil {
		log.Fatalln(err)
	}
	if err := validate.RegisterValidation("should_be_future", shouldBeFuture); err != nil {
		log.Fatalln(err)
	}
}

type ValidationErrorsResponse []*ValidationErrorResponse

func (v ValidationErrorsResponse) Error() string {
	return ErrValidationFailed.Error()
}

type ValidationErrorResponse struct {
	FailedField string `json:"failed_field,omitempty"`
	Condition   string `json:"condition"`
	ActualValue string `json:"actual_value,omitempty"`
}

//Validate validates a struct
func Validate(iface interface{}) error {
	err := validate.Struct(iface)
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

// shouldBeFuture validates if the current field is time.Time and is after time.Now().
func shouldBeFuture(fl validator.FieldLevel) bool {
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
