package request

import "github.com/shopspring/decimal"

// DepositRequest godoc
type DepositRequest struct {
	AccountID string          `json:"accountId" binding:"required,uuid"`
	Amount    decimal.Decimal `json:"amount" binding:"required"`
}

// WithdrawRequest godoc
type WithdrawRequest struct {
	AccountID string          `json:"accountId" binding:"required,uuid"`
	Amount    decimal.Decimal `json:"amount" binding:"required"`
}

// TransferRequest godoc
type TransferRequest struct {
	FromAccountID string          `json:"fromAccountId" binding:"required,uuid"`
	ToAccountID   string          `json:"toAccountId" binding:"required,uuid"`
	Amount        decimal.Decimal `json:"amount" binding:"required"`
}
