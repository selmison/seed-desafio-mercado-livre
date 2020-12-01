package mercadolivre

import (
	"log"
	"reflect"
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
	if err := validate.RegisterValidation("should_be_future", ShouldBeFuture); err != nil {
		log.Fatalln(err)
	}
}

// ShouldBeFuture is the validation function for validating if the current field
// is time.Time and is after time.Now().
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
