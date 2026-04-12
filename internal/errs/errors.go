package errs

import (
	"net/http"

	pkgerrs "github.com/sirawong/simple-banking-api/pkg/errs"
)

var (
	// Auth
	ErrInvalidPassword = pkgerrs.Sentinel("Invalid email or password", http.StatusUnauthorized)

	// User
	ErrUserNotFound  = pkgerrs.Sentinel("User not found", http.StatusNotFound)
	ErrDuplicateUser = pkgerrs.Sentinel("User already exists", http.StatusConflict)

	// Account
	ErrAccountNotFound  = pkgerrs.Sentinel("Account not found", http.StatusNotFound)
	ErrDuplicateAccount = pkgerrs.Sentinel("Account already exists", http.StatusConflict)

	// Transaction
	ErrInsufficientBalance = pkgerrs.Sentinel("Insufficient balance", http.StatusUnprocessableEntity)
	ErrSameAccount         = pkgerrs.Sentinel("From and to account must be different", http.StatusBadRequest)

	// Authorization
	ErrForbidden = pkgerrs.Sentinel("Forbidden", http.StatusForbidden)
)
