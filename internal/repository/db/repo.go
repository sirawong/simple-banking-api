package db

import (
	"context"

	"github.com/sirawong/simple-banking-api/internal/domain"
)

// Tx is an opaque handle to a database transaction.
// Implementations type-assert to the concrete driver type (e.g. *gorm.DB).
type Tx = any

// TxManager abstracts starting and committing/rolling-back a database transaction.
type TxManager interface {
	RunInTx(ctx context.Context, fn func(tx Tx) error) error
}

// AccountRepository defines persistence operations for Account entities.
type AccountRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Account, error)
	FindByUserID(ctx context.Context, userID string) ([]*domain.Account, error)
	FindByIDForUpdate(ctx context.Context, tx Tx, id string) (*domain.Account, error)
	Create(ctx context.Context, account *domain.Account) error
	Update(ctx context.Context, tx Tx, account *domain.Account) error
	Delete(ctx context.Context, id string) error
}

// UserRepository defines persistence operations for User entities.
type UserRepository interface {
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
}

// TransactionRepository defines persistence operations for Transaction entities.
type TransactionRepository interface {
	Create(ctx context.Context, tx Tx, transaction *domain.Transaction) error
	FindByAccountID(ctx context.Context, accountID string, page, limit int) ([]*domain.Transaction, int64, error)
}
