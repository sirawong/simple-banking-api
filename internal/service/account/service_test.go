package account

import (
	"context"
	"testing"

	"github.com/google/uuid"
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
	created := &entity.Account{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), Currency: "THB"}
	s.ar.EXPECT().Create(mock.Anything, mock.AnythingOfType("*entity.Account")).Return(created, nil)

	acc, err := s.svc.CreateAccount(context.Background(), "00000000-0000-0000-0000-000000000001", "THB")
	s.NoError(err)
	s.Equal(created.ID, acc.ID)
}

func (s *AccountServiceSuite) TestCreateAccount_DuplicateAccount() {
	s.ar.EXPECT().Create(mock.Anything, mock.AnythingOfType("*entity.Account")).Return(nil, errs.ErrDuplicateAccount)

	_, err := s.svc.CreateAccount(context.Background(), "00000000-0000-0000-0000-000000000001", "THB")
	s.ErrorIs(err, errs.ErrDuplicateAccount)
}

func (s *AccountServiceSuite) TestGetBalance_Success() {
	ownerID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	s.ar.EXPECT().FindByAccountNumber(mock.Anything, "acc-1").Return(&entity.Account{
		UserID:  ownerID,
		Balance: decimal.NewFromFloat(100.00),
	}, nil)
	s.cr.EXPECT().Get(mock.Anything, mock.Anything).Return("", assert.AnError)
	s.cr.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	got, err := s.svc.GetBalance(context.Background(), ownerID.String(), "acc-1")
	s.NoError(err)
	s.True(decimal.NewFromFloat(100.00).Equal(got))
}

func (s *AccountServiceSuite) TestGetBalance_AccountNotFound() {
	s.ar.EXPECT().FindByAccountNumber(mock.Anything, "acc-1").Return(nil, errs.ErrAccountNotFound)

	_, err := s.svc.GetBalance(context.Background(), "user-1", "acc-1")
	s.ErrorIs(err, errs.ErrAccountNotFound)
}

func (s *AccountServiceSuite) TestCreateAccount_InvalidUserID() {
	_, err := s.svc.CreateAccount(context.Background(), "not-a-uuid", "THB")
	s.Error(err)
}

func (s *AccountServiceSuite) TestGetBalance_Forbidden() {
	ownerID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	otherID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	s.ar.EXPECT().FindByAccountNumber(mock.Anything, "acc-1").Return(&entity.Account{
		UserID:  ownerID,
		Balance: decimal.NewFromFloat(100),
	}, nil)

	_, err := s.svc.GetBalance(context.Background(), otherID.String(), "acc-1")
	s.ErrorIs(err, errs.ErrForbidden)
}

func (s *AccountServiceSuite) TestGetBalance_CacheHit() {
	ownerID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	s.ar.EXPECT().FindByAccountNumber(mock.Anything, "acc-1").Return(&entity.Account{
		UserID:  ownerID,
		Balance: decimal.NewFromFloat(100),
	}, nil)
	s.cr.EXPECT().Get(mock.Anything, mock.Anything).Return("200", nil)

	got, err := s.svc.GetBalance(context.Background(), ownerID.String(), "acc-1")
	s.NoError(err)
	s.True(decimal.NewFromFloat(200).Equal(got))
}

func (s *AccountServiceSuite) TestListAccounts_Success() {
	userID := "00000000-0000-0000-0000-000000000001"
	accounts := []*entity.Account{
		{AccountNumber: "1111111111", Currency: "THB"},
		{AccountNumber: "2222222222", Currency: "USD"},
	}
	s.ar.EXPECT().FindByUserID(mock.Anything, userID).Return(accounts, nil)

	result, err := s.svc.ListAccounts(context.Background(), userID)
	s.NoError(err)
	s.Len(result, 2)
}

func (s *AccountServiceSuite) TestListAccounts_Empty() {
	userID := "00000000-0000-0000-0000-000000000002"
	s.ar.EXPECT().FindByUserID(mock.Anything, userID).Return([]*entity.Account{}, nil)

	result, err := s.svc.ListAccounts(context.Background(), userID)
	s.NoError(err)
	s.Empty(result)
}
