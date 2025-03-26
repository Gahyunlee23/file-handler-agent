package errors

import (
	"errors"
	"fmt"
)

var (
	ErrBadRequest         = errors.New("bad request")
	ErrNotFound           = errors.New("resource not found")
	ErrInvalidParameter   = errors.New("invalid parameter")
	ErrAgentNotFound      = errors.New("agent not found")
	ErrActionNotSupported = errors.New("action not supported")
	ErrFileNotFound       = errors.New("file not found")
	ErrExecutionFailed    = errors.New("execution failed")
)

// WithMessage additional message return together
func WithMessage(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}

// IsType check the error type
func IsType(err, target error) bool {
	return errors.Is(err, target)
}
