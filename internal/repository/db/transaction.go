package db

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sirawong/simple-banking-api/internal/domain/entity"
	"github.com/sirawong/simple-banking-api/internal/repository/db/model"
)

type transactionRepository struct {
	db *gorm.DB
}

// @wire:set(name=RepositorySet)
func ProvideTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(ctx context.Context, tx Tx, transaction *entity.Transaction) error {
	gormTx := tx.(*gorm.DB)
	if transaction.ID == uuid.Nil {
		transaction.ID = uuid.New()
	}
	return gormTx.WithContext(ctx).Create(transaction).Error
}

func (r *transactionRepository) FindByAccountID(ctx context.Context, accountID string, page, limit int) ([]*entity.Transaction, int64, error) {
	var transactions *model.Transactions
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Transaction{}).
		Where("from_account_id = ? OR to_account_id = ?", accountID, accountID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&transactions).Error; err != nil {
		return nil, 0, err
	}

	return transactions.ToEntities(), total, nil
}
