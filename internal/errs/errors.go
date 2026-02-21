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
	Reason  string
	Hint    string
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
	reason, hint := DefaultReasonHint(code)
	return &AppError{Code: code, Reason: reason, Hint: hint, Message: message}
}

func Wrap(code int, message string, err error) error {
	reason, hint := DefaultReasonHint(code)
	return &AppError{Code: code, Reason: reason, Hint: hint, Message: message, Err: err}
}

func NewDetailed(code int, reason, hint, message string) error {
	return &AppError{Code: code, Reason: reason, Hint: hint, Message: message}
}

func WrapDetailed(code int, reason, hint, message string, err error) error {
	return &AppError{Code: code, Reason: reason, Hint: hint, Message: message, Err: err}
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

func DefaultReasonHint(code int) (string, string) {
	switch code {
	case ExitValidation:
		return "validation_error", "Verify required arguments and flag values."
	case ExitNotFound:
		return "not_found", "Confirm the note/path/key exists in the selected vault."
	case ExitConfig:
		return "config_error", "Check --vault/--config values and local config files."
	default:
		return "runtime_error", "Retry with --json for structured output and inspect the failure envelope."
	}
}
