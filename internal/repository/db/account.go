package db

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/google/uuid"

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

func (r *accountRepository) FindByUserID(ctx context.Context, userID string) ([]*entity.Account, error) {
	var accounts model.Accounts
	err := dbFromCtx(ctx, r.db).
		Scopes(notDeleted).
		Where("user_id = ?", userID).
		Find(&accounts).Error
	if err != nil {
		return nil, pkgerrs.ErrInternal.WithError(err)
	}
	return accounts.ToEntities(), nil
}

func (r *accountRepository) FindByAccountNumber(ctx context.Context, accountNumber string) (*entity.Account, error) {
	var account model.Account
	err := dbFromCtx(ctx, r.db).
		Scopes(notDeleted).
		Where("account_number = ?", accountNumber).
		First(&account).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrAccountNotFound
		}
		return nil, pkgerrs.ErrInternal.WithError(err)
	}
	return account.ToDomain(), nil
}

func (r *accountRepository) FindByAccountNumberForUpdate(ctx context.Context, accountNumber string) (*entity.Account, error) {
	if err := requireTx(ctx); err != nil {
		return nil, err
	}

	var account model.Account
	err := dbFromCtx(ctx, r.db).
		Scopes(notDeleted).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("account_number = ?", accountNumber).
		First(&account).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrAccountNotFound
		}
		return nil, pkgerrs.ErrInternal.WithError(err)
	}
	return account.ToDomain(), nil
}

func (r *accountRepository) Create(ctx context.Context, account *entity.Account) (*entity.Account, error) {
	a := model.FromEntityAccount(account)
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if err := dbFromCtx(ctx, r.db).Create(a).Error; err != nil {
		if isDuplicateError(err) {
			return nil, errs.ErrDuplicateAccount
		}
		return nil, pkgerrs.ErrInternal.WithError(err)
	}
	return a.ToDomain(), nil
}

func (r *accountRepository) Update(ctx context.Context, account *entity.Account) error {
	if err := requireTx(ctx); err != nil {
		return err
	}

	a := model.FromEntityAccount(account)
	if a == nil {
		return pkgerrs.ErrBadRequest.New("invalid account")
	}
	if err := dbFromCtx(ctx, r.db).Save(a).Error; err != nil {
		return pkgerrs.ErrInternal.WithError(err)
	}
	return nil
}
