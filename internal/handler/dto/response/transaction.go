package response

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sirawong/simple-banking-api/internal/domain/constrant"
	"github.com/sirawong/simple-banking-api/internal/domain/entity"
)

type TransactionResponse struct {
	ID                uuid.UUID                   `json:"id"`
	FromAccountNumber *string                     `json:"fromAccountNumber,omitempty"`
	ToAccountNumber   *string                     `json:"toAccountNumber,omitempty"`
	Amount            decimal.Decimal             `json:"amount"`
	Type              constrant.TransactionType   `json:"type"`
	Status            constrant.TransactionStatus `json:"status"`
	CreatedAt         time.Time                   `json:"createdAt"`
}

func FromEntityTransaction(t *entity.Transaction) *TransactionResponse {
	if t == nil {
		return nil
	}
	var (
		fromAccountNumber *string
		toAccountNumber   *string
	)
	if t.FromAccount != nil {
		fromAccountNumber = &t.FromAccount.AccountNumber
	}
	if t.ToAccount != nil {
		toAccountNumber = &t.ToAccount.AccountNumber
	}
	return &TransactionResponse{
		ID:                t.ID,
		FromAccountNumber: fromAccountNumber,
		ToAccountNumber:   toAccountNumber,
		Amount:            t.Amount,
		Type:              t.Type,
		Status:            t.Status,
		CreatedAt:         t.CreatedAt,
	}
}

func FromEntityTransactions(txs []*entity.Transaction) []*TransactionResponse {
	result := make([]*TransactionResponse, len(txs))
	for i, tx := range txs {
		result[i] = FromEntityTransaction(tx)
	}
	return result
}

type TransactionListResponse struct {
	Transactions []*TransactionResponse `json:"transactions"`
	Total        int64                  `json:"total"`
	Page         int                    `json:"page"`
	Limit        int                    `json:"limit"`
}
