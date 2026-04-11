package account

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/sirawong/simple-banking-api/internal/domain/entity"
	"github.com/sirawong/simple-banking-api/internal/errs"
	cachemock "github.com/sirawong/simple-banking-api/internal/repository/cache/mock"
	dbmock "github.com/sirawong/simple-banking-api/internal/repository/db/mock"
)

type AccountServiceSuite struct {
	suite.Suite
	ar  *dbmock.MockAccountRepository
	cr  *cachemock.MockRepository
	svc Service
}

func (s *AccountServiceSuite) SetupTest() {
	s.ar = dbmock.NewMockAccountRepository(s.T())
	s.cr = cachemock.NewMockRepository(s.T())
	s.svc = ProvideService(s.ar, s.cr)
}

func TestAccountServiceSuite(t *testing.T) {
	suite.Run(t, new(AccountServiceSuite))
}

func (s *AccountServiceSuite) TestCreateAccount_Success() {
	s.ar.On("Create", mock.Anything, mock.AnythingOfType("*entity.Account")).Return(nil)

	acc, err := s.svc.CreateAccount(context.Background(), "00000000-0000-0000-0000-000000000001", "THB")
	s.NoError(err)
	s.NotNil(acc)
}

func (s *AccountServiceSuite) TestCreateAccount_DuplicateAccount() {
	s.ar.On("Create", mock.Anything, mock.AnythingOfType("*entity.Account")).Return(errs.ErrDuplicateAccount)

	_, err := s.svc.CreateAccount(context.Background(), "00000000-0000-0000-0000-000000000001", "THB")
	s.ErrorIs(err, errs.ErrDuplicateAccount)
}

func (s *AccountServiceSuite) TestGetBalance_Success() {
	s.cr.On("Get", mock.Anything, mock.Anything).Return("", assert.AnError)
	s.ar.On("FindByID", mock.Anything, "acc-1").Return(&entity.Account{
		Balance: decimal.NewFromFloat(100.00),
	}, nil)
	s.cr.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	got, err := s.svc.GetBalance(context.Background(), "acc-1")
	s.NoError(err)
	s.True(decimal.NewFromFloat(100.00).Equal(got))
}

func (s *AccountServiceSuite) TestGetBalance_AccountNotFound() {
	s.cr.On("Get", mock.Anything, mock.Anything).Return("", assert.AnError)
	s.ar.On("FindByID", mock.Anything, "acc-1").Return(nil, errs.ErrAccountNotFound)

	_, err := s.svc.GetBalance(context.Background(), "acc-1")
	s.ErrorIs(err, errs.ErrAccountNotFound)
}
