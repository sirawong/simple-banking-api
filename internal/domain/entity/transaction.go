package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sirawong/simple-banking-api/internal/domain/constrant"
)

type Transaction struct {
	ID            uuid.UUID
	FromAccountID *uuid.UUID
	ToAccountID   uuid.UUID
	Amount        decimal.Decimal
	Type          constrant.TransactionType
	Status        constrant.TransactionStatus
	Note          *string
	CreatedAt     time.Time
	FromAccount   *Account
	ToAccount     Account
}
