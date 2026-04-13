package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sirawong/simple-banking-api/internal/domain/constrant"
	"github.com/sirawong/simple-banking-api/internal/domain/entity"
)

type Transaction struct {
	ID        uuid.UUID                   `gorm:"type:uuid;primaryKey"`
	Amount    decimal.Decimal             `gorm:"type:decimal(20,2);not null"`
	Type      constrant.TransactionType   `gorm:"type:varchar(20);not null"`
	Status    constrant.TransactionStatus `gorm:"type:varchar(20);not null;default:pending"`
	CreatedAt time.Time

	FromAccountID *uuid.UUID `gorm:"type:uuid;index"`
	FromAccount   *Account   `gorm:"foreignKey:FromAccountID"`
	ToAccountID   *uuid.UUID `gorm:"type:uuid;index"`
	ToAccount     *Account   `gorm:"foreignKey:ToAccountID"`
}

func (Transaction) TableName() string {
	return "transactions"
}

func (t *Transaction) ToEntity() *entity.Transaction {
	if t == nil {
		return nil
	}
	return &entity.Transaction{
		ID:            t.ID,
		FromAccountID: t.FromAccountID,
		ToAccountID:   t.ToAccountID,
		Amount:        t.Amount,
		Type:          t.Type,
		Status:        t.Status,
		CreatedAt:     t.CreatedAt,
		FromAccount:   t.FromAccount.ToDomain(),
		ToAccount:     t.ToAccount.ToDomain(),
	}
}

type Transactions []Transaction

func (t *Transactions) ToEntities() []*entity.Transaction {
	if t == nil {
		return nil
	}
	transactions := make([]*entity.Transaction, len(*t))
	for i, transaction := range *t {
		transactions[i] = transaction.ToEntity()
	}
	return transactions
}

func FromEntityTransaction(t *entity.Transaction) *Transaction {
	if t == nil {
		return nil
	}
	return &Transaction{
		ID:            t.ID,
		FromAccountID: t.FromAccountID,
		ToAccountID:   t.ToAccountID,
		Amount:        t.Amount,
		Type:          t.Type,
		Status:        t.Status,
		CreatedAt:     t.CreatedAt,
		FromAccount:   FromEntityAccount(t.FromAccount),
		ToAccount:     FromEntityAccount(t.ToAccount),
	}
}
