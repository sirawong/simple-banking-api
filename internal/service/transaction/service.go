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
	account, err := s.accountRepo.FindByAccountNumber(ctx, accountNumber)
	if err != nil {
		return nil, errs.ErrAccountNotFound
	}
	if account.UserID.String() != userID {
		return nil, errs.ErrForbidden.New("account %s does not belong to the authenticated user", accountNumber)
	}

	tx, err := s.applyAndRecord(ctx, account, accountNumber, account.Balance.Add(amount), &entity.Transaction{
		ToAccountID: &account.ID,
		Amount:      amount,
		Type:        constrant.TransactionTypeDeposit,
		Status:      constrant.TransactionStatusSuccess,
	})
	if err != nil {
		return nil, err
	}
	tx.ToAccount = account
	return tx, nil
}

func (s *service) Withdraw(ctx context.Context, userID, accountNumber string, amount decimal.Decimal) (*entity.Transaction, error) {
	account, err := s.accountRepo.FindByAccountNumber(ctx, accountNumber)
	if err != nil {
		return nil, err
	}
	if account.UserID.String() != userID {
		return nil, errs.ErrForbidden.New("account %s does not belong to the authenticated user", accountNumber)
	}
	if account.Balance.LessThan(amount) {
		return nil, errs.ErrInsufficientBalance.New("balance %s is less than requested amount %s", account.Balance, amount)
	}

	tx, err := s.applyAndRecord(ctx, account, accountNumber, account.Balance.Sub(amount), &entity.Transaction{
		FromAccountID: &account.ID,
		Amount:        amount,
		Type:          constrant.TransactionTypeWithdraw,
		Status:        constrant.TransactionStatusSuccess,
	})
	if err != nil {
		return nil, err
	}
	tx.FromAccount = account
	return tx, nil
}

// applyAndRecord sets newBalance on the account, persists the update and the transaction record
// within a single DB transaction, then refreshes the balance cache.
func (s *service) applyAndRecord(
	ctx context.Context,
	account *entity.Account,
	accountNumber string,
	newBalance decimal.Decimal,
	record *entity.Transaction,
) (*entity.Transaction, error) {
	_ = s.cache.Delete(ctx, cacheBalanceKey(accountNumber))

	var tx *entity.Transaction
	err := s.txManager.Transaction(ctx, func(ctx context.Context) error {
		account.Balance = newBalance
		if err := s.accountRepo.Update(ctx, account); err != nil {
			return err
		}
		created, err := s.txRepo.Create(ctx, record)
		if err != nil {
			return err
		}
		tx = created
		return nil
	})
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, cacheBalanceKey(accountNumber), account.Balance.String(), 60*time.Second)
	return tx, nil
}

func (s *service) Transfer(ctx context.Context, userID, fromAccountNumber, toAccountNumber string, amount decimal.Decimal) (*entity.Transaction, error) {
	if fromAccountNumber == toAccountNumber {
		return nil, errs.ErrSameAccount
	}

	_ = s.cache.Delete(ctx, cacheBalanceKey(fromAccountNumber))
	_ = s.cache.Delete(ctx, cacheBalanceKey(toAccountNumber))

	var (
		tx   *entity.Transaction
		from *entity.Account
		to   *entity.Account
	)
	err := s.txManager.Transaction(ctx, func(ctx context.Context) error {
		var err error
		from, err = s.accountRepo.FindByAccountNumberForUpdate(ctx, fromAccountNumber)
		if err != nil {
			return errs.ErrAccountNotFound
		}
		if from.UserID.String() != userID {
			return errs.ErrForbidden.New("account %s does not belong to the authenticated user", fromAccountNumber)
		}
		to, err = s.accountRepo.FindByAccountNumberForUpdate(ctx, toAccountNumber)
		if err != nil {
			return errs.ErrAccountNotFound
		}

		if from.Balance.LessThan(amount) {
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
