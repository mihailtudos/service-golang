package validate

import (
	"encoding/json"
	"errors"
)

// ErrInvalidID occures when an ID is not in a valid form.
var ErrInvalidID = errors.New("ID is not in its proper form")

// ErrInvalidEmail occures when an email is not in a valid form.
var ErrInvalidEmail = errors.New("email address is not valid")

// ErrorResponse is used when there is a validation error.
type ErrorResponse = struct {
	Error  string `json:"error"`
	Fields string `json:"fields,omitempty"`
}

// RequestError is used to pass an error during the request through the
// application with web specific context.
type RequestError struct {
	Err    error
	Status int
	Fields error
}

// NewRequestError wraps a provided error with an HTTP status code. This
// function should be used when handlers encounter expected errors.
func NewRequestError(err error, status int) *RequestError {
	return &RequestError{err, status, nil}
}

// Error implements the error interface. It uses the default message of the
// wrapped error. This is what will be shown in the services' logs.
func (re *RequestError) Error() string {
	return re.Err.Error()
}

// FieldError is used to indicate an error with a specific request field.
type FieldError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

// FieldErrors represents a collection of field errors.
type FieldErrors []FieldError

// Error implements the error interface.
func (fe FieldErrors) Error() string {
	d, err := json.Marshal(fe)
	if err != nil {
		return err.Error()
	}
	return string(d)
}

// Cause iterates through all the wrapped errors until the root
// error is reached.
func Cause(err error) error {
	root := err
	for {
		unwrapped := errors.Unwrap(root)
		if unwrapped == nil {
			return root
		}

		root = unwrapped
	}
}
