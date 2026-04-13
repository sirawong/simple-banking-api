package db

import (
	"context"

	"github.com/google/uuid"

	adapterdb "github.com/sirawong/simple-banking-api/internal/adapter/postgres"
	"github.com/sirawong/simple-banking-api/internal/domain/entity"
	"github.com/sirawong/simple-banking-api/internal/repository/db/model"
	pkgerrs "github.com/sirawong/simple-banking-api/pkg/errs"
)

type transactionRepository struct {
	db *adapterdb.DB
}

// @wire:set(name=RepositorySet)
func ProvideTransactionRepository(db *adapterdb.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(ctx context.Context, transaction *entity.Transaction) (*entity.Transaction, error) {
	if err := requireTx(ctx); err != nil {
		return nil, err
	}
	tx := model.FromEntityTransaction(transaction)
	if tx == nil {
		return nil, pkgerrs.ErrBadRequest
	}
	if tx.ID == uuid.Nil {
		tx.ID = uuid.New()
	}
	if err := dbFromCtx(ctx, r.db).Create(tx).Error; err != nil {
		return nil, pkgerrs.ErrInternal.WithError(err)
	}
	return tx.ToEntity(), nil
}

func (r *transactionRepository) FindByAccountID(ctx context.Context, accountID string, page, limit int) ([]*entity.Transaction, int64, error) {
	var tx model.Transactions
	var total int64

	query := r.db.WithContext(ctx).
		Model(&model.Transaction{}).
		Where("from_account_id = ? OR to_account_id = ?", accountID, accountID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Preload("FromAccount").
		Preload("ToAccount").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&tx).Error
	if err != nil {
		return nil, 0, err
	}

	return tx.ToEntities(), total, nil
}
