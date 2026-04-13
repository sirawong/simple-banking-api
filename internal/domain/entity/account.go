package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Account struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	AccountNumber string
	Balance       decimal.Decimal
	Currency      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
	User          User
}
