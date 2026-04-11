package db

import (
	"context"

	"gorm.io/gorm"
)

type gormTxManager struct {
	db *gorm.DB
}

// @wire:set(name=RepositorySet)
func ProvideTxManager(db *gorm.DB) TxManager {
	return &gormTxManager{db: db}
}

func (m *gormTxManager) RunInTx(ctx context.Context, fn func(tx Tx) error) error {
	return m.db.WithContext(ctx).Transaction(func(gormTx *gorm.DB) error {
		return fn(gormTx)
	})
}
