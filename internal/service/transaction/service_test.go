package transaction

import (
	"context"
	"testing"

	"github.com/google/uuid"
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
	txm *dbmock.MockTxManager
	ar  *dbmock.MockAccountRepository
	tr  *dbmock.MockTransactionRepository
	cr  *cachemock.MockRepository
	svc Service
}

func (s *TransactionServiceSuite) SetupTest() {
	s.txm = dbmock.NewMockTxManager(s.T())
	s.ar = dbmock.NewMockAccountRepository(s.T())
	s.tr = dbmock.NewMockTransactionRepository(s.T())
	s.cr = cachemock.NewMockRepository(s.T())
	s.svc = ProvideService(s.txm, s.ar, s.tr, s.cr)
}

func TestTransactionServiceSuite(t *testing.T) {
	suite.Run(t, new(TransactionServiceSuite))
}

// execTx makes the mock TxManager execute the closure, simulating a real transaction.
func (s *TransactionServiceSuite) execTx() {
	s.txm.EXPECT().Transaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
}

func (s *TransactionServiceSuite) TestTransfer_SameAccount() {
	_, err := s.svc.Transfer(context.Background(), "user-1", "same-id", "same-id", decimal.NewFromFloat(100))
	s.ErrorIs(err, errs.ErrSameAccount)
}

func (s *TransactionServiceSuite) TestTransfer_CurrencyMismatch() {
	ownerID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	s.cr.EXPECT().Delete(mock.Anything, mock.Anything).Return(nil).Times(2)
	s.execTx()
	s.ar.EXPECT().FindByAccountNumberForUpdate(mock.Anything, "acc-thb").Return(&entity.Account{
		UserID:   ownerID,
		Currency: "THB",
		Balance:  decimal.NewFromFloat(500),
	}, nil)
	s.ar.EXPECT().FindByAccountNumberForUpdate(mock.Anything, "acc-usd").Return(&entity.Account{
		Currency: "USD",
	}, nil)

	_, err := s.svc.Transfer(context.Background(), ownerID.String(), "acc-thb", "acc-usd", decimal.NewFromFloat(100))
	s.ErrorIs(err, errs.ErrCurrencyMismatch)
}

func (s *TransactionServiceSuite) TestDeposit_AccountNotFound() {
	s.cr.EXPECT().Delete(mock.Anything, mock.Anything).Return(nil)
	s.execTx()
	s.ar.EXPECT().FindByAccountNumberForUpdate(mock.Anything, "acc-missing").Return(nil, errs.ErrAccountNotFound)

	_, err := s.svc.Deposit(context.Background(), "user-1", "acc-missing", decimal.NewFromFloat(100))
	s.ErrorIs(err, errs.ErrAccountNotFound)
}

func (s *TransactionServiceSuite) TestDeposit_Forbidden() {
	ownerID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	otherID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	s.cr.EXPECT().Delete(mock.Anything, mock.Anything).Return(nil)
	s.execTx()
	s.ar.EXPECT().FindByAccountNumberForUpdate(mock.Anything, "acc-1").Return(&entity.Account{UserID: ownerID}, nil)

	_, err := s.svc.Deposit(context.Background(), otherID.String(), "acc-1", decimal.NewFromFloat(100))
	s.ErrorIs(err, errs.ErrForbidden)
}

func (s *TransactionServiceSuite) TestWithdraw_AccountNotFound() {
	s.cr.EXPECT().Delete(mock.Anything, mock.Anything).Return(nil)
	s.execTx()
	s.ar.EXPECT().FindByAccountNumberForUpdate(mock.Anything, "acc-missing").Return(nil, errs.ErrAccountNotFound)

	_, err := s.svc.Withdraw(context.Background(), "user-1", "acc-missing", decimal.NewFromFloat(100))
	s.ErrorIs(err, errs.ErrAccountNotFound)
}

func (s *TransactionServiceSuite) TestWithdraw_Forbidden() {
	ownerID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	otherID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	s.cr.EXPECT().Delete(mock.Anything, mock.Anything).Return(nil)
	s.execTx()
	s.ar.EXPECT().FindByAccountNumberForUpdate(mock.Anything, "acc-1").Return(&entity.Account{
		UserID:  ownerID,
		Balance: decimal.NewFromFloat(500),
	}, nil)

	_, err := s.svc.Withdraw(context.Background(), otherID.String(), "acc-1", decimal.NewFromFloat(100))
	s.ErrorIs(err, errs.ErrForbidden)
}

func (s *TransactionServiceSuite) TestWithdraw_InsufficientBalance() {
	ownerID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	s.cr.EXPECT().Delete(mock.Anything, mock.Anything).Return(nil)
	s.execTx()
	s.ar.EXPECT().FindByAccountNumberForUpdate(mock.Anything, "acc-1").Return(&entity.Account{
		UserID:  ownerID,
		Balance: decimal.NewFromFloat(50),
	}, nil)

	_, err := s.svc.Withdraw(context.Background(), ownerID.String(), "acc-1", decimal.NewFromFloat(100))
	s.ErrorIs(err, errs.ErrInsufficientBalance)
}

func (s *TransactionServiceSuite) TestListByAccount_ReturnsPaginatedResults() {
	ownerID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	s.ar.EXPECT().FindByAccountNumber(mock.Anything, "1234567890").Return(&entity.Account{ID: ownerID, UserID: ownerID}, nil)
	txs := []*entity.Transaction{
		{Type: constrant.TransactionTypeDeposit, Amount: decimal.NewFromFloat(100)},
		{Type: constrant.TransactionTypeWithdraw, Amount: decimal.NewFromFloat(50)},
	}
	s.tr.On("FindByAccountID", mock.Anything, ownerID.String(), 1, 10).Return(txs, int64(2), nil)

	result, total, err := s.svc.ListByAccount(context.Background(), ownerID.String(), "1234567890", 1, 10)
	s.NoError(err)
	s.Equal(int64(2), total)
	s.Len(result, 2)
}

func (s *TransactionServiceSuite) TestListByAccount_EmptyResult() {
	ownerID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	s.ar.EXPECT().FindByAccountNumber(mock.Anything, "0987654321").Return(&entity.Account{ID: ownerID, UserID: ownerID}, nil)
	s.tr.On("FindByAccountID", mock.Anything, ownerID.String(), 1, 10).Return([]*entity.Transaction{}, int64(0), nil)

	result, total, err := s.svc.ListByAccount(context.Background(), ownerID.String(), "0987654321", 1, 10)
	s.NoError(err)
	s.Equal(int64(0), total)
	s.Empty(result)
}

func (s *TransactionServiceSuite) TestListByAccount_AccountNotFound() {
	s.ar.EXPECT().FindByAccountNumber(mock.Anything, "acc-missing").Return(nil, errs.ErrAccountNotFound)

	_, _, err := s.svc.ListByAccount(context.Background(), "user-1", "acc-missing", 1, 10)
	s.ErrorIs(err, errs.ErrAccountNotFound)
}

func (s *TransactionServiceSuite) TestListByAccount_Forbidden() {
	ownerID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	otherID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	s.ar.EXPECT().FindByAccountNumber(mock.Anything, "acc-1").Return(&entity.Account{
		ID:     ownerID,
		UserID: ownerID,
	}, nil)

	_, _, err := s.svc.ListByAccount(context.Background(), otherID.String(), "acc-1", 1, 10)
	s.ErrorIs(err, errs.ErrForbidden)
}
