// Package validate provides validation utilities.
package validate

import (
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/google/uuid"
	"reflect"
	"strings"
)

// validate holds the settings and caches for validating request struct values.
var validate *validator.Validate

// translator is a cache of locale and translation information.
var translator ut.Translator

func init() {
	// Initialize the validator.
	validate = validator.New()

	// Create a translator to use for validation errors.
	translator, _ = ut.New(en.New(), en.New()).GetTranslator("en")

	// Register the english error messages for use.
	en_translations.RegisterDefaultTranslations(validate, translator)

	// Use json tag names for errors instead of Go struct names
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}

		return name
	})
}

// Check validates the provided model against it's declared tags.
func Check(val any) error {
	if err := validate.Struct(val); err != nil {
		// Use type assertion to get the underlying value of the error
		verrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return err
		}

		var fields FieldErrors
		for _, verror := range verrors {
			field := FieldError{
				Field: verror.Field(),
				Error: verror.Translate(translator),
			}
			fields = append(fields, field)
		}
		return fields
	}

	return nil
}

// GenerateID creates a unique ID.
func GenerateID() string {
	return uuid.NewString()
}

// CheckID validates that the format of an id is valid.
func CheckID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidID
	}

	return nil
}

func Email(email string) error {
	if err := validate.Var(email, "required,email"); err != nil {
		return ErrInvalidEmail
	}

	return nil
}
