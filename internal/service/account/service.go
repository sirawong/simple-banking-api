package account

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sirawong/simple-banking-api/internal/domain/constrant"
	"github.com/sirawong/simple-banking-api/internal/domain/entity"
	"github.com/sirawong/simple-banking-api/internal/errs"
	cacherepo "github.com/sirawong/simple-banking-api/internal/repository/cache"
	dbrepo "github.com/sirawong/simple-banking-api/internal/repository/db"
	pkgerrs "github.com/sirawong/simple-banking-api/pkg/errs"
)

type Service interface {
	CreateAccount(ctx context.Context, userID, currency string) (*entity.Account, error)
	GetBalance(ctx context.Context, accountID string) (decimal.Decimal, error)
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
		AccountNumber: generateAccountNumber(),
		Currency:      currency,
		Balance:       decimal.Zero,
	}
	if err := s.accountRepo.Create(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *service) GetBalance(ctx context.Context, accountID string) (decimal.Decimal, error) {
	cacheKey := constrant.CacheKeyPrefixAccountBalance + accountID
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		if d, err := decimal.NewFromString(cached); err == nil {
			return d, nil
		}
	}

	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return decimal.Zero, errs.ErrAccountNotFound
	}

	_ = s.cache.Set(ctx, cacheKey, account.Balance.String(), 60*time.Second)
	return account.Balance, nil
}

func generateAccountNumber() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%010d", r.Int63n(9000000000)+1000000000)
}
