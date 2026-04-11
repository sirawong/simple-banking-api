package dto

import "github.com/shopspring/decimal"

// DepositRequest godoc
type DepositRequest struct {
	AccountID string          `json:"account_id" binding:"required,uuid"`
	Amount    decimal.Decimal `json:"amount" binding:"required"`
}

// WithdrawRequest godoc
type WithdrawRequest struct {
	AccountID string          `json:"account_id" binding:"required,uuid"`
	Amount    decimal.Decimal `json:"amount" binding:"required"`
}

// TransferRequest godoc
type TransferRequest struct {
	FromAccountID string          `json:"from_account_id" binding:"required,uuid"`
	ToAccountID   string          `json:"to_account_id" binding:"required,uuid"`
	Amount        decimal.Decimal `json:"amount" binding:"required"`
}
