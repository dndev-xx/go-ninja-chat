package validator

import (
	"strings"

	"github.com/go-playground/validator/v10"
	optsGenValidator "github.com/kazhuravlev/options-gen/pkg/validator"
)

var Validator = validator.New()

func init() {
	optsGenValidator.Set(Validator)
	registerCustomValidations(Validator)
}

func registerCustomValidations(v *validator.Validate) {
	v.RegisterValidation("hostname_port", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		parts := strings.Split(value, ":")
		if len(parts) != 2 {
			return false
		}
		port := parts[1]
		return port != ""
	})
}
