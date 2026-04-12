package account

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sirawong/simple-banking-api/internal/domain/constrant"
	"github.com/sirawong/simple-banking-api/internal/domain/entity"
	"github.com/sirawong/simple-banking-api/internal/errs"
	cacherepo "github.com/sirawong/simple-banking-api/internal/repository/cache"
	dbrepo "github.com/sirawong/simple-banking-api/internal/repository/db"
	"github.com/sirawong/simple-banking-api/internal/utils"
	pkgerrs "github.com/sirawong/simple-banking-api/pkg/errs"
	"github.com/sirawong/simple-banking-api/pkg/logger"
)

type Service interface {
	CreateAccount(ctx context.Context, userID, currency string) (*entity.Account, error)
	GetBalance(ctx context.Context, userID, accountNumber string) (decimal.Decimal, error)
	ListAccounts(ctx context.Context, userID string) ([]*entity.Account, error)
}

type service struct {
	accountRepo dbrepo.AccountRepository
	cache       cacherepo.Repository
}

// @wire:set(name=ServiceSet)
func ProvideService(accountRepo dbrepo.AccountRepository, cache cacherepo.Repository) Service {
	return &service{accountRepo: accountRepo, cache: cache}
}

func (s *service) CreateAccount(ctx context.Context, userID, currency string) (*entity.Account, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, pkgerrs.ErrBadRequest.New("invalid user_id: %s", userID)
	}
	account := &entity.Account{
		UserID:        uid,
		AccountNumber: utils.GenerateAccountNumber(),
		Currency:      currency,
		Balance:       decimal.Zero,
	}
	return s.accountRepo.Create(ctx, account)
}

func (s *service) GetBalance(ctx context.Context, userID, accountNumber string) (decimal.Decimal, error) {
	account, err := s.accountRepo.FindByAccountNumber(ctx, accountNumber)
	if err != nil {
		return decimal.Zero, errs.ErrAccountNotFound
	}
	if account.UserID.String() != userID {
		logger.Warn("get balance forbidden", "userID", userID, "accountNumber", accountNumber)
		return decimal.Zero, errs.ErrForbidden.New("account %s does not belong to the authenticated user", accountNumber)
	}

	cacheKey := constrant.CacheKeyPrefixAccountBalance + accountNumber
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		if d, err := decimal.NewFromString(cached); err == nil {
			return d, nil
		}
	}

	_ = s.cache.Set(ctx, cacheKey, account.Balance.String(), 60*time.Second)
	return account.Balance, nil
}

func (s *service) ListAccounts(ctx context.Context, userID string) ([]*entity.Account, error) {
	return s.accountRepo.FindByUserID(ctx, userID)
}
