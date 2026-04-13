package db

import (
	"context"

	"gorm.io/gorm"

	adapterdb "github.com/sirawong/simple-banking-api/internal/adapter/postgres"
)

// txContextKey is an unexported key for storing a gorm transaction in context.
type txContextKey struct{}

func dbFromCtx(ctx context.Context, base *adapterdb.DB) *gorm.DB {
	if tx, ok := ctx.Value(txContextKey{}).(*gorm.DB); ok && tx != nil {
		return tx.WithContext(ctx)
	}
	return base.WithContext(ctx)
}

type gormTxManager struct {
	db *adapterdb.DB
}

// @wire:set(name=RepositorySet)
func ProvideTxManager(db *adapterdb.DB) TxManager {
	return &gormTxManager{db: db}
}

func (m *gormTxManager) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return dbFromCtx(ctx, m.db).Transaction(func(gormTx *gorm.DB) error {
		return fn(context.WithValue(ctx, txContextKey{}, gormTx))
	})
}
