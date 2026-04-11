package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Account struct {
	ID            uuid.UUID       `gorm:"type:uuid;primaryKey"           json:"id"`
	UserID        uuid.UUID       `gorm:"type:uuid;index;not null"       json:"user_id"`
	AccountNumber string          `gorm:"uniqueIndex;not null"           json:"account_number"`
	Balance       decimal.Decimal `gorm:"type:decimal(20,2);not null;default:0" json:"balance"`
	Currency      string          `gorm:"not null;default:THB"           json:"currency"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	DeletedAt     *time.Time      `gorm:"index"                          json:"-"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}
