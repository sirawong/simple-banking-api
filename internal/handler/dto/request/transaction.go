package request

import "github.com/shopspring/decimal"

// DepositRequest godoc
type DepositRequest struct {
	AccountNumber string          `json:"accountNumber" binding:"required"`
	Amount        decimal.Decimal `json:"amount" binding:"required"`
}

// WithdrawRequest godoc
type WithdrawRequest struct {
	AccountNumber string          `json:"accountNumber" binding:"required"`
	Amount        decimal.Decimal `json:"amount" binding:"required"`
}

// TransferRequest godoc
type TransferRequest struct {
	FromAccountNumber string          `json:"fromAccountNumber" binding:"required"`
	ToAccountNumber   string          `json:"toAccountNumber" binding:"required"`
	Amount            decimal.Decimal `json:"amount" binding:"required"`
}
