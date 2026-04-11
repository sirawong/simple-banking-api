package response

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sirawong/simple-banking-api/internal/domain/entity"
)

type AccountResponse struct {
	ID            uuid.UUID       `json:"id"`
	UserID        uuid.UUID       `json:"userId"`
	AccountNumber string          `json:"accountNumber"`
	Balance       decimal.Decimal `json:"balance"`
	Currency      string          `json:"currency"`
	CreatedAt     time.Time       `json:"createdAt"`
}

func FromEntityAccount(a *entity.Account) *AccountResponse {
	if a == nil {
		return nil
	}
	return &AccountResponse{
		ID:            a.ID,
		UserID:        a.UserID,
		AccountNumber: a.AccountNumber,
		Balance:       a.Balance,
		Currency:      a.Currency,
		CreatedAt:     a.CreatedAt,
	}
}

type BalanceResponse struct {
	AccountID string          `json:"accountId"`
	Balance   decimal.Decimal `json:"balance"`
}
