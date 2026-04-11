package account

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

// --- Tests ---

func TestAccountService_CreateAccount(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*mockAccountRepo, *mockCacheRepo)
		wantErr   error
	}{
		{
			name: "success",
			setupMock: func(ar *mockAccountRepo, cr *mockCacheRepo) {
				ar.On("Create", mock.Anything, mock.AnythingOfType("*domain.Account")).Return(nil)
			},
		},
		{
			name: "duplicate account number",
			setupMock: func(ar *mockAccountRepo, cr *mockCacheRepo) {
				ar.On("Create", mock.Anything, mock.AnythingOfType("*domain.Account")).Return(errs.ErrDuplicateAccount)
			},
			wantErr: errs.ErrDuplicateAccount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ar := &mockAccountRepo{}
			cr := &mockCacheRepo{}
			tt.setupMock(ar, cr)

			svc := ProvideService(ar, cr)
			_, err := svc.CreateAccount(context.Background(), "00000000-0000-0000-0000-000000000001", "THB")
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
			ar.AssertExpectations(t)
		})
	}
}

func TestAccountService_GetBalance(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*mockAccountRepo, *mockCacheRepo)
		want      decimal.Decimal
		wantErr   error
	}{
		{
			name: "success",
			setupMock: func(ar *mockAccountRepo, cr *mockCacheRepo) {
				cr.On("Get", mock.Anything, mock.Anything).Return("", assert.AnError)
				ar.On("FindByID", mock.Anything, "acc-1").Return(&domain.Account{
					Balance: decimal.NewFromFloat(100.00),
				}, nil)
				cr.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			want: decimal.NewFromFloat(100.00),
		},
		{
			name: "account not found",
			setupMock: func(ar *mockAccountRepo, cr *mockCacheRepo) {
				cr.On("Get", mock.Anything, mock.Anything).Return("", assert.AnError)
				ar.On("FindByID", mock.Anything, "acc-1").Return(nil, errs.ErrAccountNotFound)
			},
			wantErr: errs.ErrAccountNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ar := &mockAccountRepo{}
			cr := &mockCacheRepo{}
			tt.setupMock(ar, cr)

			svc := ProvideService(ar, cr)
			got, err := svc.GetBalance(context.Background(), "acc-1")
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.want.Equal(got))
			}
			ar.AssertExpectations(t)
			cr.AssertExpectations(t)
		})
	}
}
