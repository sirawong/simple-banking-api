package transaction

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/sirawong/simple-banking-api/internal/domain"
	"github.com/sirawong/simple-banking-api/internal/errs"
	dbrepo "github.com/sirawong/simple-banking-api/internal/repository/db"
)

// --- Mocks ---

type mockAccountRepo struct{ mock.Mock }

func (m *mockAccountRepo) FindByID(ctx context.Context, id string) (*domain.Account, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Account), args.Error(1)
}
func (m *mockAccountRepo) FindByUserID(ctx context.Context, userID string) ([]*domain.Account, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*domain.Account), args.Error(1)
}
func (m *mockAccountRepo) FindByIDForUpdate(ctx context.Context, tx dbrepo.Tx, id string) (*domain.Account, error) {
	args := m.Called(ctx, tx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Account), args.Error(1)
}
func (m *mockAccountRepo) Create(ctx context.Context, acct *domain.Account) error {
	return m.Called(ctx, acct).Error(0)
}
func (m *mockAccountRepo) Update(ctx context.Context, tx dbrepo.Tx, acct *domain.Account) error {
	return m.Called(ctx, tx, acct).Error(0)
}
func (m *mockAccountRepo) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

type mockCacheRepo struct{ mock.Mock }

func (m *mockCacheRepo) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}
func (m *mockCacheRepo) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return m.Called(ctx, key, value, ttl).Error(0)
}
func (m *mockCacheRepo) Delete(ctx context.Context, key string) error {
	return m.Called(ctx, key).Error(0)
}

type mockTxRepo struct{ mock.Mock }

func (m *mockTxRepo) Create(ctx context.Context, tx dbrepo.Tx, t *domain.Transaction) error {
	return m.Called(ctx, tx, t).Error(0)
}
func (m *mockTxRepo) FindByAccountID(ctx context.Context, accountID string, page, limit int) ([]*domain.Transaction, int64, error) {
	args := m.Called(ctx, accountID, page, limit)
	return args.Get(0).([]*domain.Transaction), args.Get(1).(int64), args.Error(2)
}

// newTxService creates a transaction service with a nil TxManager for unit tests.
// Only early-return error paths are tested here; happy paths need a real DB.
func newTxService(ar *mockAccountRepo, tr *mockTxRepo, cr *mockCacheRepo) Service {
	return ProvideService(nil, ar, tr, cr)
}

func TestTransactionService_Transfer_SameAccount(t *testing.T) {
	svc := newTxService(&mockAccountRepo{}, &mockTxRepo{}, &mockCacheRepo{})
	_, err := svc.Transfer(context.Background(), "same-id", "same-id", decimal.NewFromFloat(100))
	assert.ErrorIs(t, err, errs.ErrSameAccount)
}

func TestTransactionService_Deposit_AccountNotFound(t *testing.T) {
	ar := &mockAccountRepo{}
	ar.On("FindByID", mock.Anything, "acc-missing").Return(nil, errs.ErrAccountNotFound)

	svc := newTxService(ar, &mockTxRepo{}, &mockCacheRepo{})
	_, err := svc.Deposit(context.Background(), "acc-missing", decimal.NewFromFloat(100))
	assert.ErrorIs(t, err, errs.ErrAccountNotFound)
	ar.AssertExpectations(t)
}

func TestTransactionService_Withdraw_AccountNotFound(t *testing.T) {
	ar := &mockAccountRepo{}
	ar.On("FindByID", mock.Anything, "acc-missing").Return(nil, errs.ErrAccountNotFound)

	svc := newTxService(ar, &mockTxRepo{}, &mockCacheRepo{})
	_, err := svc.Withdraw(context.Background(), "acc-missing", decimal.NewFromFloat(100))
	assert.ErrorIs(t, err, errs.ErrAccountNotFound)
	ar.AssertExpectations(t)
}

func TestTransactionService_Withdraw_InsufficientBalance(t *testing.T) {
	ar := &mockAccountRepo{}
	ar.On("FindByID", mock.Anything, "acc-1").Return(&domain.Account{
		Balance: decimal.NewFromFloat(50),
	}, nil)

	svc := newTxService(ar, &mockTxRepo{}, &mockCacheRepo{})
	_, err := svc.Withdraw(context.Background(), "acc-1", decimal.NewFromFloat(100))
	assert.ErrorIs(t, err, errs.ErrInsufficientBalance)
	ar.AssertExpectations(t)
}

func TestTransactionService_ListByAccount(t *testing.T) {
	tests := []struct {
		name      string
		accountID string
		page      int
		limit     int
		setupMock func(*mockTxRepo)
		wantTotal int64
		wantLen   int
		wantErr   bool
	}{
		{
			name:      "returns paginated results",
			accountID: "acc-1",
			page:      1,
			limit:     10,
			setupMock: func(tr *mockTxRepo) {
				txs := []*domain.Transaction{
					{Type: domain.TransactionTypeDeposit, Amount: decimal.NewFromFloat(100)},
					{Type: domain.TransactionTypeWithdraw, Amount: decimal.NewFromFloat(50)},
				}
				tr.On("FindByAccountID", mock.Anything, "acc-1", 1, 10).Return(txs, int64(2), nil)
			},
			wantTotal: 2,
			wantLen:   2,
		},
		{
			name:      "empty result",
			accountID: "acc-2",
			page:      1,
			limit:     10,
			setupMock: func(tr *mockTxRepo) {
				tr.On("FindByAccountID", mock.Anything, "acc-2", 1, 10).Return([]*domain.Transaction{}, int64(0), nil)
			},
			wantTotal: 0,
			wantLen:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &mockTxRepo{}
			tt.setupMock(tr)

			svc := newTxService(&mockAccountRepo{}, tr, &mockCacheRepo{})
			txs, total, err := svc.ListByAccount(context.Background(), tt.accountID, tt.page, tt.limit)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTotal, total)
				assert.Len(t, txs, tt.wantLen)
			}
			tr.AssertExpectations(t)
		})
	}
}
