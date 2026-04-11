package response

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sirawong/simple-banking-api/internal/domain/constrant"
	"github.com/sirawong/simple-banking-api/internal/domain/entity"
)

type TransactionResponse struct {
	ID            uuid.UUID                   `json:"id"`
	FromAccountID *uuid.UUID                  `json:"fromAccountId,omitempty"`
	ToAccountID   uuid.UUID                   `json:"toAccountId"`
	Amount        decimal.Decimal             `json:"amount"`
	Type          constrant.TransactionType   `json:"type"`
	Status        constrant.TransactionStatus `json:"status"`
	Note          *string                     `json:"note,omitempty"`
	CreatedAt     time.Time                   `json:"createdAt"`
}

func FromEntityTransaction(t *entity.Transaction) *TransactionResponse {
	if t == nil {
		return nil
	}
	return &TransactionResponse{
		ID:            t.ID,
		FromAccountID: t.FromAccountID,
		ToAccountID:   t.ToAccountID,
		Amount:        t.Amount,
		Type:          t.Type,
		Status:        t.Status,
		Note:          t.Note,
		CreatedAt:     t.CreatedAt,
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
