package errs

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError is a structured error that carries an HTTP status code and optional detail.
type AppError struct {
	Message        string `json:"error_message"`
	Detail         any    `json:"detail,omitempty"`
	HttpStatusCode int    `json:"-"`
	Unwrap         error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Unwrap == nil {
		return fmt.Sprintf("message: %s", e.Message)
	}
	return fmt.Sprintf("message: %s, unwrap: %s", e.Message, e.Unwrap.Error())
}

func (e *AppError) Is(target error) bool {
	var t *AppError
	if !errors.As(target, &t) {
		return false
	}
	return e.Message == t.Message && e.HttpStatusCode == t.HttpStatusCode
}

// Wrap returns a new AppError with an underlying error and a formatted detail message.
func (e *AppError) Wrap(err error, detailFormat string, args ...any) *AppError {
	return &AppError{
		Message:        e.Message,
		HttpStatusCode: e.HttpStatusCode,
		Detail:         fmt.Sprintf(detailFormat, args...),
		Unwrap:         err,
	}
}

// WithError returns a new AppError carrying the given error as both detail and unwrap target.
func (e *AppError) WithError(err error) *AppError {
	return &AppError{
		Message:        e.Message,
		HttpStatusCode: e.HttpStatusCode,
		Detail:         err.Error(),
		Unwrap:         err,
	}
}

// WithDetails returns a new AppError with additional structured detail.
func (e *AppError) WithDetails(details any) *AppError {
	return &AppError{
		Message:        e.Message,
		HttpStatusCode: e.HttpStatusCode,
		Detail:         details,
	}
}

// New returns a new AppError with a formatted detail message.
func (e *AppError) New(detailFormat string, args ...any) *AppError {
	return &AppError{
		Message:        e.Message,
		HttpStatusCode: e.HttpStatusCode,
		Detail:         fmt.Sprintf(detailFormat, args...),
	}
}

// Sentinel creates a base sentinel error with a fixed message and HTTP status code.
func Sentinel(message string, httpStatusCode int) *AppError {
	return &AppError{
		Message:        message,
		HttpStatusCode: httpStatusCode,
	}
}

var (
	ErrInternal     = Sentinel("Internal Server Error", http.StatusInternalServerError)
	ErrBadRequest   = Sentinel("Invalid request", http.StatusBadRequest)
	ErrUnauthorized = Sentinel("Unauthorized", http.StatusUnauthorized)
	ErrInvalidToken = Sentinel("Invalid or expired token", http.StatusUnauthorized)
)
