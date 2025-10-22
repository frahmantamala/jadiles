package common

import (
	"errors"
	"strings"

	"github.com/frahmantamala/jadiles/internal"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator"
)

var (
	validate   *validator.Validate
	translator ut.Translator
)

func ValidateStruct(s any) error {
	err := validate.Struct(s)
	if err != nil {
		return handleValidateStructErr(err)
	}
	return nil
}

func handleValidateStructErr(err error) error {
	errMessage := err.Error()

	var validationErrors validator.ValidationErrors
	if ok := errors.As(err, &validationErrors); ok {
		validationErrorMessages := make([]string, 0, len(validationErrors))

		for _, validationError := range validationErrors {
			validationErrorMessage := validationError.Translate(translator)
			validationErrorMessages = append(validationErrorMessages, validationErrorMessage)
		}

		errMessage = strings.Join(validationErrorMessages, ", ")
	}

	return internal.NewValidationError(errMessage)
}
