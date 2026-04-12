package db

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	adapterdb "github.com/sirawong/simple-banking-api/internal/adapter/postgres"
	"github.com/sirawong/simple-banking-api/internal/domain/entity"
	"github.com/sirawong/simple-banking-api/internal/errs"
	"github.com/sirawong/simple-banking-api/internal/repository/db/model"
	pkgerrs "github.com/sirawong/simple-banking-api/pkg/errs"
)

type accountRepository struct {
	db *adapterdb.DB
}

// @wire:set(name=RepositorySet)
func ProvideAccountRepository(db *adapterdb.DB) AccountRepository {
	return &accountRepository{db: db}
}

func (r *accountRepository) FindByID(ctx context.Context, id string) (*entity.Account, error) {
	var account model.Account
	if err := dbFromCtx(ctx, r.db).Where("deleted_at IS NULL").First(&account, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrAccountNotFound
		}
		return nil, pkgerrs.ErrInternal.Wrap(err, "Internal error")
	}
	return account.ToDomain(), nil
}

func (r *accountRepository) FindByUserID(ctx context.Context, userID string) ([]*entity.Account, error) {
	var accounts model.Accounts
	if err := dbFromCtx(ctx, r.db).Where("user_id = ? AND deleted_at IS NULL", userID).Find(&accounts).Error; err != nil {
		return nil, pkgerrs.ErrInternal.Wrap(err, "Internal error")
	}
	return accounts.ToEntities(), nil
}

func (r *accountRepository) FindByIDForUpdate(ctx context.Context, id string) (*entity.Account, error) {
	if _, ok := ctx.Value(txContextKey{}).(*gorm.DB); !ok {
		return nil, pkgerrs.ErrInternal.New("FindByIDForUpdate must be called within a transaction")
	}
	var account model.Account
	if err := dbFromCtx(ctx, r.db).Set("gorm:query_option", "FOR UPDATE").
		Where("deleted_at IS NULL").First(&account, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrAccountNotFound
		}
		return nil, pkgerrs.ErrInternal.Wrap(err, "Internal error")
	}
	return account.ToDomain(), nil
}

func (r *accountRepository) Create(ctx context.Context, account *entity.Account) error {
	if account.ID == uuid.Nil {
		account.ID = uuid.New()
	}
	m := model.FromEntityAccount(account)
	if err := dbFromCtx(ctx, r.db).Create(m).Error; err != nil {
		if isDuplicateError(err) {
			return errs.ErrDuplicateAccount
		}
		return pkgerrs.ErrInternal.Wrap(err, "Internal error")
	}
	return nil
}

func (r *accountRepository) Update(ctx context.Context, account *entity.Account) error {
	m := model.FromEntityAccount(account)
	if m == nil {
		return pkgerrs.ErrBadRequest.New("invalid account")
	}
	return dbFromCtx(ctx, r.db).Save(m).Error
}

func (r *accountRepository) Delete(ctx context.Context, id string) error {
	return dbFromCtx(ctx, r.db).
		Model(&model.Account{}).
		Where("id = ?", id).
		Update("deleted_at", gorm.Expr("NOW()")).Error
}
