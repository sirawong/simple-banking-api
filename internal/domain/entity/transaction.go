package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sirawong/simple-banking-api/internal/domain/constrant"
)

type Transaction struct {
	ID            uuid.UUID
	Amount        decimal.Decimal
	Type          constrant.TransactionType
	Status        constrant.TransactionStatus
	Note          *string
	CreatedAt     time.Time
	FromAccountID *uuid.UUID
	FromAccount   *Account
	ToAccountID   *uuid.UUID
	ToAccount     *Account
}
