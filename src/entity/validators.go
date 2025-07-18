package entity

import (
	"github.com/go-playground/validator/v10"
)

type IError struct {
	Field string
	Tag   string
	Value string
}

var validate *validator.Validate = validator.New()

func GetStructError(err error) []IError {
	var errors []IError

	if err == nil {
		return errors
	}

	// if _, ok := err.(*validator.InvalidValidationError); ok {
	for _, err := range err.(validator.ValidationErrors) {
		errors = append(errors, IError{
			Field: err.Field(),
			Tag:   err.Tag(),
			Value: err.Value().(string),
		})
	}
	// }

	return errors
}
