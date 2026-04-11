package transaction

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/sirawong/simple-banking-api/internal/domain/constrant"
	"github.com/sirawong/simple-banking-api/internal/domain/entity"
	"github.com/sirawong/simple-banking-api/internal/errs"
	cachemock "github.com/sirawong/simple-banking-api/internal/repository/cache/mock"
	dbmock "github.com/sirawong/simple-banking-api/internal/repository/db/mock"
)

type TransactionServiceSuite struct {
	suite.Suite
	ar  *dbmock.MockAccountRepository
	tr  *dbmock.MockTransactionRepository
	cr  *cachemock.MockRepository
	svc Service
}

func (s *TransactionServiceSuite) SetupTest() {
	s.ar = dbmock.NewMockAccountRepository(s.T())
	s.tr = dbmock.NewMockTransactionRepository(s.T())
	s.cr = cachemock.NewMockRepository(s.T())
	// nil TxManager: only early-return error paths are tested here; happy paths need a real DB.
	s.svc = ProvideService(nil, s.ar, s.tr, s.cr)
}

func TestTransactionServiceSuite(t *testing.T) {
	suite.Run(t, new(TransactionServiceSuite))
}

func (s *TransactionServiceSuite) TestTransfer_SameAccount() {
	_, err := s.svc.Transfer(context.Background(), "same-id", "same-id", decimal.NewFromFloat(100))
	s.ErrorIs(err, errs.ErrSameAccount)
}

func (s *TransactionServiceSuite) TestDeposit_AccountNotFound() {
	s.ar.On("FindByID", mock.Anything, "acc-missing").Return(nil, errs.ErrAccountNotFound)

	_, err := s.svc.Deposit(context.Background(), "acc-missing", decimal.NewFromFloat(100))
	s.ErrorIs(err, errs.ErrAccountNotFound)
}

func (s *TransactionServiceSuite) TestWithdraw_AccountNotFound() {
	s.ar.On("FindByID", mock.Anything, "acc-missing").Return(nil, errs.ErrAccountNotFound)

	_, err := s.svc.Withdraw(context.Background(), "acc-missing", decimal.NewFromFloat(100))
	s.ErrorIs(err, errs.ErrAccountNotFound)
}

func (s *TransactionServiceSuite) TestWithdraw_InsufficientBalance() {
	s.ar.On("FindByID", mock.Anything, "acc-1").Return(&entity.Account{
		Balance: decimal.NewFromFloat(50),
	}, nil)

	_, err := s.svc.Withdraw(context.Background(), "acc-1", decimal.NewFromFloat(100))
	s.ErrorIs(err, errs.ErrInsufficientBalance)
}

func (s *TransactionServiceSuite) TestListByAccount_ReturnsPaginatedResults() {
	txs := []*entity.Transaction{
		{Type: constrant.TransactionTypeDeposit, Amount: decimal.NewFromFloat(100)},
		{Type: constrant.TransactionTypeWithdraw, Amount: decimal.NewFromFloat(50)},
	}
	s.tr.On("FindByAccountID", mock.Anything, "acc-1", 1, 10).Return(txs, int64(2), nil)

	result, total, err := s.svc.ListByAccount(context.Background(), "acc-1", 1, 10)
	s.NoError(err)
	s.Equal(int64(2), total)
	s.Len(result, 2)
}

func (s *TransactionServiceSuite) TestListByAccount_EmptyResult() {
	s.tr.On("FindByAccountID", mock.Anything, "acc-2", 1, 10).Return([]*entity.Transaction{}, int64(0), nil)

	result, total, err := s.svc.ListByAccount(context.Background(), "acc-2", 1, 10)
	s.NoError(err)
	s.Equal(int64(0), total)
	s.Empty(result)
}
