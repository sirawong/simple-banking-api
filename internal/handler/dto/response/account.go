package response

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sirawong/simple-banking-api/internal/domain/entity"
)

type AccountResponse struct {
	UserID        uuid.UUID       `json:"userId"`
	AccountNumber string          `json:"accountNumber"`
	Balance       decimal.Decimal `json:"balance"`
	Currency      string          `json:"currency"`
}

func FromEntityAccount(a *entity.Account) *AccountResponse {
	if a == nil {
		return nil
	}
	return &AccountResponse{
		UserID:        a.UserID,
		AccountNumber: a.AccountNumber,
		Balance:       a.Balance,
		Currency:      a.Currency,
	}
}

func FromEntityAccounts(accounts []*entity.Account) []*AccountResponse {
	res := make([]*AccountResponse, 0, len(accounts))
	for _, a := range accounts {
		res = append(res, FromEntityAccount(a))
	}
	return res
}

type BalanceResponse struct {
	AccountNumber string          `json:"accountNumber"`
	Balance       decimal.Decimal `json:"balance"`
}
