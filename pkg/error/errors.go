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

	ErrUnsupportedFormat = errors.New("unsupported format")
	ErrDirectoryCreation = errors.New("directory creation failed")
	ErrProcessTimeout    = errors.New("process timed out")
)

// FormatError represents an error with a specific format
type FormatError struct {
	Format string
	Err    error
}

// Error implements the error interface
func (e *FormatError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err.Error(), e.Format)
}

// Unwrap returns the wrapped error
func (e *FormatError) Unwrap() error {
	return e.Err
}

// NewUnsupportedFormatError creates a new format error
func NewUnsupportedFormatError(format string) error {
	return &FormatError{
		Format: format,
		Err:    ErrUnsupportedFormat,
	}
}

// WithMessage additional message return together
func WithMessage(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}

// IsType check the error type
func IsType(err, target error) bool {
	return errors.Is(err, target)
}

// IsUnsupportedFormat checks if the error is an unsupported format error
func IsUnsupportedFormat(err error) bool {
	return errors.Is(err, ErrUnsupportedFormat)
}

// GetFormatFromError extracts the format from a FormatError if present
func GetFormatFromError(err error) (string, bool) {
	var formatErr *FormatError
	if errors.As(err, &formatErr) {
		return formatErr.Format, true
	}
	return "", false
}
