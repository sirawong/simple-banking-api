package transaction

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/sirawong/simple-banking-api/internal/domain/constrant"
	"github.com/sirawong/simple-banking-api/internal/domain/entity"
	"github.com/sirawong/simple-banking-api/internal/errs"
	cacherepo "github.com/sirawong/simple-banking-api/internal/repository/cache"
	dbrepo "github.com/sirawong/simple-banking-api/internal/repository/db"
)

type Service interface {
	Deposit(ctx context.Context, accountID string, amount decimal.Decimal) (*entity.Transaction, error)
	Withdraw(ctx context.Context, accountID string, amount decimal.Decimal) (*entity.Transaction, error)
	Transfer(ctx context.Context, fromAccountID, toAccountID string, amount decimal.Decimal) (*entity.Transaction, error)
	ListByAccount(ctx context.Context, accountID string, page, limit int) ([]*entity.Transaction, int64, error)
}

type service struct {
	txManager   dbrepo.TxManager
	accountRepo dbrepo.AccountRepository
	txRepo      dbrepo.TransactionRepository
	cache       cacherepo.Repository
}

// @wire:set(name=ServiceSet)
func ProvideService(
	txManager dbrepo.TxManager,
	accountRepo dbrepo.AccountRepository,
	txRepo dbrepo.TransactionRepository,
	cache cacherepo.Repository,
) Service {
	return &service{
		txManager:   txManager,
		accountRepo: accountRepo,
		txRepo:      txRepo,
		cache:       cache,
	}
}

func (s *service) Deposit(ctx context.Context, accountID string, amount decimal.Decimal) (*entity.Transaction, error) {
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return nil, errs.ErrAccountNotFound
	}

	_ = s.cache.Delete(ctx, cacheBalanceKey(accountID))

	var tx *entity.Transaction
	err = s.txManager.RunInTx(ctx, func(dbTx dbrepo.Tx) error {
		account.Balance = account.Balance.Add(amount)
		if err := s.accountRepo.Update(ctx, dbTx, account); err != nil {
			return err
		}

		toID := account.ID
		tx = &entity.Transaction{
			ToAccountID: toID,
			Amount:      amount,
			Type:        constrant.TransactionTypeDeposit,
			Status:      constrant.TransactionStatusSuccess,
		}
		return s.txRepo.Create(ctx, dbTx, tx)
	})
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, cacheBalanceKey(accountID), account.Balance.String(), 60*time.Second)
	return tx, nil
}

func (s *service) Withdraw(ctx context.Context, accountID string, amount decimal.Decimal) (*entity.Transaction, error) {
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return nil, errs.ErrAccountNotFound
	}
	if account.Balance.LessThan(amount) {
		return nil, errs.ErrInsufficientBalance
	}

	_ = s.cache.Delete(ctx, cacheBalanceKey(accountID))

	var tx *entity.Transaction
	err = s.txManager.RunInTx(ctx, func(dbTx dbrepo.Tx) error {
		account.Balance = account.Balance.Sub(amount)
		if err := s.accountRepo.Update(ctx, dbTx, account); err != nil {
			return err
		}

		fromID := account.ID
		tx = &entity.Transaction{
			FromAccountID: &fromID,
			ToAccountID:   account.ID,
			Amount:        amount,
			Type:          constrant.TransactionTypeWithdraw,
			Status:        constrant.TransactionStatusSuccess,
		}
		return s.txRepo.Create(ctx, dbTx, tx)
	})
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, cacheBalanceKey(accountID), account.Balance.String(), 60*time.Second)
	return tx, nil
}

func (s *service) Transfer(ctx context.Context, fromAccountID, toAccountID string, amount decimal.Decimal) (*entity.Transaction, error) {
	if fromAccountID == toAccountID {
		return nil, errs.ErrSameAccount
	}

	_ = s.cache.Delete(ctx, cacheBalanceKey(fromAccountID))
	_ = s.cache.Delete(ctx, cacheBalanceKey(toAccountID))

	var tx *entity.Transaction
	err := s.txManager.RunInTx(ctx, func(dbTx dbrepo.Tx) error {
		from, err := s.accountRepo.FindByIDForUpdate(ctx, dbTx, fromAccountID)
		if err != nil {
			return errs.ErrAccountNotFound
		}
		to, err := s.accountRepo.FindByIDForUpdate(ctx, dbTx, toAccountID)
		if err != nil {
			return errs.ErrAccountNotFound
		}

		if from.Balance.LessThan(amount) {
			return errs.ErrInsufficientBalance
		}

		from.Balance = from.Balance.Sub(amount)
		to.Balance = to.Balance.Add(amount)

		if err := s.accountRepo.Update(ctx, dbTx, from); err != nil {
			return err
		}
		if err := s.accountRepo.Update(ctx, dbTx, to); err != nil {
			return err
		}

		fromID := from.ID
		tx = &entity.Transaction{
			FromAccountID: &fromID,
			ToAccountID:   to.ID,
			Amount:        amount,
			Type:          constrant.TransactionTypeTransfer,
			Status:        constrant.TransactionStatusSuccess,
		}
		return s.txRepo.Create(ctx, dbTx, tx)
	})
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, cacheBalanceKey(fromAccountID), "invalidated", time.Second)
	_ = s.cache.Set(ctx, cacheBalanceKey(toAccountID), "invalidated", time.Second)
	return tx, nil
}

func (s *service) ListByAccount(ctx context.Context, accountID string, page, limit int) ([]*entity.Transaction, int64, error) {
	return s.txRepo.FindByAccountID(ctx, accountID, page, limit)
}

func cacheBalanceKey(accountID string) string {
	return fmt.Sprintf("account:balance:%s", accountID)
}
