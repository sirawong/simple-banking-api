package db

import (
	"context"

	"github.com/sirawong/simple-banking-api/internal/domain/entity"
)

// TxManager abstracts starting and committing/rolling-back a database transaction.
// The active transaction is propagated via context so callers do not handle it directly.
type TxManager interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// AccountRepository defines persistence operations for Account entities.
type AccountRepository interface {
	FindByUserID(ctx context.Context, userID string) ([]*entity.Account, error)
	FindByAccountNumber(ctx context.Context, accountNumber string) (*entity.Account, error)
	FindByAccountNumberForUpdate(ctx context.Context, accountNumber string) (*entity.Account, error)
	Create(ctx context.Context, account *entity.Account) (*entity.Account, error)
	Update(ctx context.Context, account *entity.Account) error
}

// UserRepository defines persistence operations for User entities.
type UserRepository interface {
	FindByID(ctx context.Context, id string) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) (*entity.User, error)
}

// TokenRepository defines persistence operations for RefreshToken entities.
type TokenRepository interface {
	Create(ctx context.Context, token *entity.RefreshToken) (*entity.RefreshToken, error)
	FindByToken(ctx context.Context, token string) (*entity.RefreshToken, error)
	DeleteByToken(ctx context.Context, token string) error
}

// TransactionRepository defines persistence operations for Transaction entities.
type TransactionRepository interface {
	Create(ctx context.Context, transaction *entity.Transaction) (*entity.Transaction, error)
	FindByAccountID(ctx context.Context, accountID string, page, limit int) ([]*entity.Transaction, int64, error)
}
