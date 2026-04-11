package db

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sirawong/simple-banking-api/internal/domain/entity"
	"github.com/sirawong/simple-banking-api/internal/errs"
	"github.com/sirawong/simple-banking-api/internal/repository/db/model"
)

type accountRepository struct {
	db *gorm.DB
}

// @wire:set(name=RepositorySet)
func ProvideAccountRepository(db *gorm.DB) AccountRepository {
	return &accountRepository{db: db}
}

func (r *accountRepository) FindByID(ctx context.Context, id string) (*entity.Account, error) {
	var account model.Account
	if err := r.db.WithContext(ctx).Where("deleted_at IS NULL").First(&account, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrAccountNotFound
		}
		return nil, err
	}
	return account.ToDomain(), nil
}

func (r *accountRepository) FindByUserID(ctx context.Context, userID string) ([]*entity.Account, error) {
	var accounts model.Accounts
	if err := r.db.WithContext(ctx).Where("user_id = ? AND deleted_at IS NULL", userID).Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts.ToEntities(), nil
}

func (r *accountRepository) FindByIDForUpdate(ctx context.Context, tx Tx, id string) (*entity.Account, error) {
	gormTx := tx.(*gorm.DB)
	var account model.Account
	if err := gormTx.WithContext(ctx).Set("gorm:query_option", "FOR UPDATE").
		Where("deleted_at IS NULL").First(&account, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrAccountNotFound
		}
		return nil, err
	}
	return account.ToDomain(), nil
}

func (r *accountRepository) Create(ctx context.Context, account *entity.Account) error {
	if account.ID == uuid.Nil {
		account.ID = uuid.New()
	}
	if err := r.db.WithContext(ctx).Create(account).Error; err != nil {
		if isDuplicateError(err) {
			return errs.ErrDuplicateAccount
		}
		return err
	}
	return nil
}

func (r *accountRepository) Update(ctx context.Context, tx Tx, account *entity.Account) error {
	gormTx := tx.(*gorm.DB)
	accountModel := model.FromEntityAccount(account)
	if accountModel == nil {
		return errs.ErrBadRequest.New("invalid account")
	}
	return gormTx.WithContext(ctx).Save(accountModel).Error
}

func (r *accountRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&model.Account{}).
		Where("id = ?", id).
		Update("deleted_at", gorm.Expr("NOW()")).Error
}
