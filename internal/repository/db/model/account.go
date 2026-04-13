package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sirawong/simple-banking-api/internal/domain/entity"
)

type Account struct {
	ID            uuid.UUID       `gorm:"type:uuid;primaryKey"`
	UserID        uuid.UUID       `gorm:"type:uuid;index;not null"`
	AccountNumber string          `gorm:"uniqueIndex;not null"`
	Balance       decimal.Decimal `gorm:"type:decimal(20,2);not null;default:0"`
	Currency      string          `gorm:"not null;default:THB"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time `gorm:"index"`

	User User `gorm:"foreignKey:UserID"`
}

func (Account) TableName() string {
	return "accounts"
}

func (a *Account) ToDomain() *entity.Account {
	if a == nil {
		return nil
	}
	return &entity.Account{
		ID:            a.ID,
		UserID:        a.UserID,
		AccountNumber: a.AccountNumber,
		Balance:       a.Balance,
		Currency:      a.Currency,
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
		DeletedAt:     a.DeletedAt,
	}
}

type Accounts []Account

func (a *Accounts) ToEntities() []*entity.Account {
	if a == nil {
		return nil
	}
	accounts := make([]*entity.Account, len(*a))
	for i, account := range *a {
		accounts[i] = account.ToDomain()
	}
	return accounts
}

func FromEntityAccount(a *entity.Account) *Account {
	if a == nil {
		return nil
	}
	return &Account{
		ID:            a.ID,
		UserID:        a.UserID,
		AccountNumber: a.AccountNumber,
		Balance:       a.Balance,
		Currency:      a.Currency,
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
		DeletedAt:     a.DeletedAt,
	}
}
