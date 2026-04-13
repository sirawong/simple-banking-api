package transaction

import (
	"context"
	"time"

	"github.com/shopspring/decimal"

	"github.com/sirawong/simple-banking-api/internal/domain/constrant"
	"github.com/sirawong/simple-banking-api/internal/domain/entity"
	"github.com/sirawong/simple-banking-api/internal/errs"
	cacherepo "github.com/sirawong/simple-banking-api/internal/repository/cache"
	dbrepo "github.com/sirawong/simple-banking-api/internal/repository/db"
	"github.com/sirawong/simple-banking-api/pkg/logger"
)

type Service interface {
	Deposit(ctx context.Context, userID, accountNumber string, amount decimal.Decimal) (*entity.Transaction, error)
	Withdraw(ctx context.Context, userID, accountNumber string, amount decimal.Decimal) (*entity.Transaction, error)
	Transfer(ctx context.Context, userID, fromAccountNumber, toAccountNumber string, amount decimal.Decimal) (*entity.Transaction, error)
	ListByAccount(ctx context.Context, userID, accountNumber string, page, limit int) ([]*entity.Transaction, int64, error)
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

func (s *service) Deposit(ctx context.Context, userID, accountNumber string, amount decimal.Decimal) (*entity.Transaction, error) {
	account, tx, err := s.applyAndRecord(ctx, accountNumber, func(account *entity.Account) (*entity.Transaction, error) {
		if account.UserID.String() != userID {
			logger.Warn("deposit forbidden", "userID", userID, "accountNumber", accountNumber)
			return nil, errs.ErrForbidden.New("account %s does not belong to the authenticated user", accountNumber)
		}
		account.Balance = account.Balance.Add(amount)
		return &entity.Transaction{
			ToAccountID: &account.ID,
			Amount:      amount,
			Type:        constrant.TransactionTypeDeposit,
			Status:      constrant.TransactionStatusSuccess,
		}, nil
	})
	if err != nil {
		return nil, err
	}
	tx.ToAccount = account
	logger.Info("deposit success", "userID", userID, "accountNumber", accountNumber, "amount", amount)
	return tx, nil
}

func (s *service) Withdraw(ctx context.Context, userID, accountNumber string, amount decimal.Decimal) (*entity.Transaction, error) {
	account, tx, err := s.applyAndRecord(ctx, accountNumber, func(account *entity.Account) (*entity.Transaction, error) {
		if account.UserID.String() != userID {
			logger.Warn("withdraw forbidden", "userID", userID, "accountNumber", accountNumber)
			return nil, errs.ErrForbidden.New("account %s does not belong to the authenticated user", accountNumber)
		}
		if account.Balance.LessThan(amount) {
			logger.Warn("withdraw insufficient balance", "userID", userID, "accountNumber", accountNumber, "balance", account.Balance, "amount", amount)
			return nil, errs.ErrInsufficientBalance.New("balance %s is less than requested amount %s", account.Balance, amount)
		}
		account.Balance = account.Balance.Sub(amount)
		return &entity.Transaction{
			FromAccountID: &account.ID,
			Amount:        amount,
			Type:          constrant.TransactionTypeWithdraw,
			Status:        constrant.TransactionStatusSuccess,
		}, nil
	})
	if err != nil {
		return nil, err
	}
	tx.FromAccount = account
	logger.Info("withdraw success", "userID", userID, "accountNumber", accountNumber, "amount", amount)
	return tx, nil
}

func (s *service) applyAndRecord(
	ctx context.Context,
	accountNumber string,
	transform func(*entity.Account) (*entity.Transaction, error),
) (*entity.Account, *entity.Transaction, error) {
	err := s.cache.Delete(ctx, cacheBalanceKey(accountNumber))
	if err != nil {
		logger.Warn("failed to delete cache", "accountNumber", accountNumber, "err", err)
	}

	var (
		account *entity.Account
		tx      *entity.Transaction
		record  *entity.Transaction
	)
	err = s.txManager.Transaction(ctx, func(ctx context.Context) error {
		account, err = s.accountRepo.FindByAccountNumberForUpdate(ctx, accountNumber)
		if err != nil {
			return err
		}

		record, err = transform(account)
		if err != nil {
			return err
		}

		if err = s.accountRepo.Update(ctx, account); err != nil {
			return err
		}
		tx, err = s.txRepo.Create(ctx, record)
		return err
	})
	if err != nil {
		return nil, nil, err
	}

	err = s.cache.Set(ctx, cacheBalanceKey(accountNumber), account.Balance.String(), 60*time.Second)
	if err != nil {
		logger.Warn("failed to set cache", "accountNumber", accountNumber, "err", err)
	}

	return account, tx, nil
}

func (s *service) Transfer(ctx context.Context, userID, fromAccountNumber, toAccountNumber string, amount decimal.Decimal) (*entity.Transaction, error) {
	if fromAccountNumber == toAccountNumber {
		return nil, errs.ErrSameAccount
	}

	err := s.cache.Delete(ctx, cacheBalanceKey(fromAccountNumber))
	if err != nil {
		logger.Warn("failed to delete cache", "accountNumber", fromAccountNumber, "err", err)
	}
	err = s.cache.Delete(ctx, cacheBalanceKey(toAccountNumber))
	if err != nil {
		logger.Warn("failed to delete cache", "accountNumber", fromAccountNumber, "err", err)
	}

	var (
		tx   *entity.Transaction
		from *entity.Account
		to   *entity.Account
	)
	err = s.txManager.Transaction(ctx, func(ctx context.Context) error {
		from, err = s.accountRepo.FindByAccountNumberForUpdate(ctx, fromAccountNumber)
		if err != nil {
			return errs.ErrAccountNotFound
		}
		if from.UserID.String() != userID {
			logger.Warn("transfer forbidden", "userID", userID, "fromAccountNumber", fromAccountNumber)
			return errs.ErrForbidden.New("account %s does not belong to the authenticated user", fromAccountNumber)
		}
		to, err = s.accountRepo.FindByAccountNumberForUpdate(ctx, toAccountNumber)
		if err != nil {
			return errs.ErrAccountNotFound
		}

		if from.Currency != to.Currency {
			return errs.ErrCurrencyMismatch.New("cannot transfer between %s and %s accounts", from.Currency, to.Currency)
		}

		if from.Balance.LessThan(amount) {
			logger.Warn("transfer insufficient balance", "userID", userID, "fromAccountNumber", fromAccountNumber, "balance", from.Balance, "amount", amount)
			return errs.ErrInsufficientBalance.New("balance %s is less than requested amount %s", from.Balance, amount)
		}

		from.Balance = from.Balance.Sub(amount)
		to.Balance = to.Balance.Add(amount)

		if err = s.accountRepo.Update(ctx, from); err != nil {
			return err
		}
		if err = s.accountRepo.Update(ctx, to); err != nil {
			return err
		}

		fromID := from.ID
		tx, err = s.txRepo.Create(ctx, &entity.Transaction{
			FromAccountID: &fromID,
			ToAccountID:   &to.ID,
			Amount:        amount,
			Type:          constrant.TransactionTypeTransfer,
			Status:        constrant.TransactionStatusSuccess,
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	tx.FromAccount = from
	tx.ToAccount = to
	logger.Info("transfer success", "userID", userID, "fromAccountNumber", fromAccountNumber, "toAccountNumber", toAccountNumber, "amount", amount)
	return tx, nil
}

func (s *service) ListByAccount(ctx context.Context, userID, accountNumber string, page, limit int) ([]*entity.Transaction, int64, error) {
	account, err := s.accountRepo.FindByAccountNumber(ctx, accountNumber)
	if err != nil {
		return nil, 0, err
	}
	if account.UserID.String() != userID {
		return nil, 0, errs.ErrForbidden.New("account %s does not belong to the authenticated user", accountNumber)
	}
	return s.txRepo.FindByAccountID(ctx, account.ID.String(), page, limit)
}

func cacheBalanceKey(accountNumber string) string {
	return constrant.CacheKeyPrefixAccountBalance + accountNumber
}
