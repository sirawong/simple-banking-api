package errs

import (
	"errors"
	"fmt"
	"net/http"
)

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
	ok := errors.As(target, &t)
	if !ok {
		return false
	}
	return e.Message == t.Message && e.HttpStatusCode == t.HttpStatusCode
}

func (e *AppError) Wrap(err error, detailFormat string, args ...any) *AppError {
	return &AppError{
		Message:        e.Message,
		HttpStatusCode: e.HttpStatusCode,
		Detail:         fmt.Sprintf(detailFormat, args...),
		Unwrap:         err,
	}
}

func (e *AppError) WithError(err error) *AppError {
	return &AppError{
		Message:        e.Message,
		HttpStatusCode: e.HttpStatusCode,
		Detail:         err.Error(),
		Unwrap:         err,
	}
}

func (e *AppError) WithDetails(details any) *AppError {
	return &AppError{
		Message:        e.Message,
		HttpStatusCode: e.HttpStatusCode,
		Detail:         details,
	}
}

func (e *AppError) New(detailFormat string, args ...any) *AppError {
	return &AppError{
		Message:        e.Message,
		HttpStatusCode: e.HttpStatusCode,
		Detail:         fmt.Sprintf(detailFormat, args...),
	}
}

var (
	// Generic errors
	ErrInternal   = &AppError{Message: "Internal Server Error", HttpStatusCode: http.StatusInternalServerError}
	ErrBadRequest = &AppError{Message: "Invalid request", HttpStatusCode: http.StatusBadRequest}

	// Account errors
	ErrAccountNotFound  = &AppError{Message: "Account not found", HttpStatusCode: http.StatusNotFound}
	ErrDuplicateAccount = &AppError{Message: "Account already exists", HttpStatusCode: http.StatusConflict}

	// User errors
	ErrUserNotFound  = &AppError{Message: "User not found", HttpStatusCode: http.StatusNotFound}
	ErrDuplicateUser = &AppError{Message: "User already exists", HttpStatusCode: http.StatusConflict}

	// Transaction errors
	ErrInsufficientBalance = &AppError{Message: "Insufficient balance", HttpStatusCode: http.StatusUnprocessableEntity}
	ErrSameAccount         = &AppError{Message: "From and to account must be different", HttpStatusCode: http.StatusBadRequest}
)
