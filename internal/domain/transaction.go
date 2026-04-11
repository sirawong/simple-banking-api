package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionType string
type TransactionStatus string

const (
	TransactionTypeDeposit  TransactionType = "deposit"
	TransactionTypeWithdraw TransactionType = "withdraw"
	TransactionTypeTransfer TransactionType = "transfer"

	TransactionStatusPending TransactionStatus = "pending"
	TransactionStatusSuccess TransactionStatus = "success"
	TransactionStatusFailed  TransactionStatus = "failed"
)

type Transaction struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`

	// from_account_id is nullable — deposits have no source account
	FromAccountID *uuid.UUID `gorm:"type:uuid;index"      json:"from_account_id,omitempty"`
	ToAccountID   uuid.UUID  `gorm:"type:uuid;index;not null" json:"to_account_id"`

	Amount decimal.Decimal   `gorm:"type:decimal(20,2);not null" json:"amount"`
	Type   TransactionType   `gorm:"type:varchar(20);not null"   json:"type"`
	Status TransactionStatus `gorm:"type:varchar(20);not null;default:pending" json:"status"`
	Note   *string           `gorm:"type:text"                   json:"note,omitempty"`

	CreatedAt time.Time `json:"created_at"`

	FromAccount *Account `gorm:"foreignKey:FromAccountID" json:"-"`
	ToAccount   Account  `gorm:"foreignKey:ToAccountID"   json:"-"`
}
