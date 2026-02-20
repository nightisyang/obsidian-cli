package errs

import (
	"errors"
	"fmt"
)

const (
	ExitOK         = 0
	ExitGeneric    = 1
	ExitValidation = 2
	ExitNotFound   = 3
	ExitConfig     = 4
)

type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code int, message string) error {
	return &AppError{Code: code, Message: message}
}

func Wrap(code int, message string, err error) error {
	return &AppError{Code: code, Message: message, Err: err}
}

func ExitCode(err error) int {
	if err == nil {
		return ExitOK
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return ExitGeneric
}
