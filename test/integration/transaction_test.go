package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/sirawong/simple-banking-api/internal/errs"
	dtores "github.com/sirawong/simple-banking-api/internal/handler/dto/response"
)

type TransactionSuite struct{ BaseSuite }

func TestTransactionSuite(t *testing.T) { suite.Run(t, new(TransactionSuite)) }

// --- Deposit ---

func (s *TransactionSuite) TestDeposit_Success() {
	user := s.seedUser("Alice", "alice@example.com", "Password@123")
	acc := s.seedAccount(user, "THB")

	w := s.POST("/api/v1/transactions/deposit", map[string]any{
		"accountNumber": acc.AccountNumber,
		"amount":        "500.00",
	}, s.tokenFor(user))

	s.Equal(http.StatusOK, w.Code)
	var body dtores.TransactionResponse
	s.decodeBody(w, &body)
	s.Equal("500", body.Amount.String())
}

func (s *TransactionSuite) TestDeposit_AccountNotFound() {
	user := s.seedUser("Alice", "alice@example.com", "Password@123")

	w := s.POST("/api/v1/transactions/deposit", map[string]any{
		"accountNumber": "9999999999",
		"amount":        "100.00",
	}, s.tokenFor(user))

	s.Equal(http.StatusNotFound, w.Code)
	s.requireErrMessage(w, errs.ErrAccountNotFound)
}

func (s *TransactionSuite) TestDeposit_ForbiddenOtherUser() {
	alice := s.seedUser("Alice", "alice@example.com", "Password@123")
	bob := s.seedUser("Bob", "bob@example.com", "Password@123")
	acc := s.seedAccount(alice, "THB")

	w := s.POST("/api/v1/transactions/deposit", map[string]any{
		"accountNumber": acc.AccountNumber,
		"amount":        "100.00",
	}, s.tokenFor(bob))

	s.Equal(http.StatusForbidden, w.Code)
	s.requireErrMessage(w, errs.ErrForbidden)
}

func (s *TransactionSuite) TestDeposit_UpdatesBalance() {
	user := s.seedUser("Alice", "alice@example.com", "Password@123")
	acc := s.seedAccountWithBalance(user, "THB", "100.00")

	s.POST("/api/v1/transactions/deposit", map[string]any{
		"accountNumber": acc.AccountNumber,
		"amount":        "250.00",
	}, s.tokenFor(user))

	w := s.GET(fmt.Sprintf("/api/v1/accounts/%s", acc.AccountNumber), s.tokenFor(user))
	s.Equal(http.StatusOK, w.Code)
	var body dtores.BalanceResponse
	s.decodeBody(w, &body)
	s.Equal("350", body.Balance.String())
}

// --- Withdraw ---

func (s *TransactionSuite) TestWithdraw_Success() {
	user := s.seedUser("Alice", "alice@example.com", "Password@123")
	acc := s.seedAccountWithBalance(user, "THB", "500.00")

	w := s.POST("/api/v1/transactions/withdraw", map[string]any{
		"accountNumber": acc.AccountNumber,
		"amount":        "200.00",
	}, s.tokenFor(user))

	s.Equal(http.StatusOK, w.Code)
	var body dtores.TransactionResponse
	s.decodeBody(w, &body)
	s.Equal("200", body.Amount.String())
}

func (s *TransactionSuite) TestWithdraw_InsufficientBalance() {
	user := s.seedUser("Alice", "alice@example.com", "Password@123")
	acc := s.seedAccountWithBalance(user, "THB", "50.00")

	w := s.POST("/api/v1/transactions/withdraw", map[string]any{
		"accountNumber": acc.AccountNumber,
		"amount":        "200.00",
	}, s.tokenFor(user))

	s.Equal(http.StatusUnprocessableEntity, w.Code)
	s.requireErrMessage(w, errs.ErrInsufficientBalance)
}

func (s *TransactionSuite) TestWithdraw_AccountNotFound() {
	user := s.seedUser("Alice", "alice@example.com", "Password@123")

	w := s.POST("/api/v1/transactions/withdraw", map[string]any{
		"accountNumber": "9999999999",
		"amount":        "100.00",
	}, s.tokenFor(user))

	s.Equal(http.StatusNotFound, w.Code)
}

func (s *TransactionSuite) TestWithdraw_ForbiddenOtherUser() {
	alice := s.seedUser("Alice", "alice@example.com", "Password@123")
	bob := s.seedUser("Bob", "bob@example.com", "Password@123")
	acc := s.seedAccountWithBalance(alice, "THB", "500.00")

	w := s.POST("/api/v1/transactions/withdraw", map[string]any{
		"accountNumber": acc.AccountNumber,
		"amount":        "100.00",
	}, s.tokenFor(bob))

	s.Equal(http.StatusForbidden, w.Code)
	s.requireErrMessage(w, errs.ErrForbidden)
}

func (s *TransactionSuite) TestWithdraw_UpdatesBalance() {
	user := s.seedUser("Alice", "alice@example.com", "Password@123")
	acc := s.seedAccountWithBalance(user, "THB", "500.00")

	s.POST("/api/v1/transactions/withdraw", map[string]any{
		"accountNumber": acc.AccountNumber,
		"amount":        "150.00",
	}, s.tokenFor(user))

	w := s.GET(fmt.Sprintf("/api/v1/accounts/%s", acc.AccountNumber), s.tokenFor(user))
	s.Equal(http.StatusOK, w.Code)
	var body dtores.BalanceResponse
	s.decodeBody(w, &body)
	s.Equal("350", body.Balance.String())
}

// --- Transfer ---

func (s *TransactionSuite) TestTransfer_Success() {
	alice := s.seedUser("Alice", "alice@example.com", "Password@123")
	bob := s.seedUser("Bob", "bob@example.com", "Password@123")
	fromAcc := s.seedAccountWithBalance(alice, "THB", "1000.00")
	toAcc := s.seedAccount(bob, "THB")

	w := s.POST("/api/v1/transactions/transfer", map[string]any{
		"fromAccountNumber": fromAcc.AccountNumber,
		"toAccountNumber":   toAcc.AccountNumber,
		"amount":            "300.00",
	}, s.tokenFor(alice))

	s.Equal(http.StatusOK, w.Code)
	var body dtores.TransactionResponse
	s.decodeBody(w, &body)
	s.Equal("300", body.Amount.String())
}

func (s *TransactionSuite) TestTransfer_UpdatesBothBalances() {
	alice := s.seedUser("Alice", "alice@example.com", "Password@123")
	bob := s.seedUser("Bob", "bob@example.com", "Password@123")
	fromAcc := s.seedAccountWithBalance(alice, "THB", "1000.00")
	toAcc := s.seedAccountWithBalance(bob, "THB", "200.00")

	s.POST("/api/v1/transactions/transfer", map[string]any{
		"fromAccountNumber": fromAcc.AccountNumber,
		"toAccountNumber":   toAcc.AccountNumber,
		"amount":            "400.00",
	}, s.tokenFor(alice))

	wFrom := s.GET(fmt.Sprintf("/api/v1/accounts/%s", fromAcc.AccountNumber), s.tokenFor(alice))
	s.Equal(http.StatusOK, wFrom.Code)
	var fromBody dtores.BalanceResponse
	s.decodeBody(wFrom, &fromBody)
	s.Equal("600", fromBody.Balance.String())

	wTo := s.GET(fmt.Sprintf("/api/v1/accounts/%s", toAcc.AccountNumber), s.tokenFor(bob))
	s.Equal(http.StatusOK, wTo.Code)
	var toBody dtores.BalanceResponse
	s.decodeBody(wTo, &toBody)
	s.Equal("600", toBody.Balance.String())
}

func (s *TransactionSuite) TestTransfer_SameAccount() {
	alice := s.seedUser("Alice", "alice@example.com", "Password@123")
	acc := s.seedAccountWithBalance(alice, "THB", "500.00")

	w := s.POST("/api/v1/transactions/transfer", map[string]any{
		"fromAccountNumber": acc.AccountNumber,
		"toAccountNumber":   acc.AccountNumber,
		"amount":            "100.00",
	}, s.tokenFor(alice))

	s.Equal(http.StatusBadRequest, w.Code)
	s.requireErrMessage(w, errs.ErrSameAccount)
}

func (s *TransactionSuite) TestTransfer_InsufficientBalance() {
	alice := s.seedUser("Alice", "alice@example.com", "Password@123")
	bob := s.seedUser("Bob", "bob@example.com", "Password@123")
	fromAcc := s.seedAccountWithBalance(alice, "THB", "50.00")
	toAcc := s.seedAccount(bob, "THB")

	w := s.POST("/api/v1/transactions/transfer", map[string]any{
		"fromAccountNumber": fromAcc.AccountNumber,
		"toAccountNumber":   toAcc.AccountNumber,
		"amount":            "500.00",
	}, s.tokenFor(alice))

	s.Equal(http.StatusUnprocessableEntity, w.Code)
	s.requireErrMessage(w, errs.ErrInsufficientBalance)
}

func (s *TransactionSuite) TestTransfer_ForbiddenFromAccount() {
	alice := s.seedUser("Alice", "alice@example.com", "Password@123")
	bob := s.seedUser("Bob", "bob@example.com", "Password@123")
	aliceAcc := s.seedAccountWithBalance(alice, "THB", "500.00")
	bobAcc := s.seedAccount(bob, "THB")

	w := s.POST("/api/v1/transactions/transfer", map[string]any{
		"fromAccountNumber": aliceAcc.AccountNumber,
		"toAccountNumber":   bobAcc.AccountNumber,
		"amount":            "100.00",
	}, s.tokenFor(bob))

	s.Equal(http.StatusForbidden, w.Code)
}

func (s *TransactionSuite) TestTransfer_ToAccountNotFound() {
	alice := s.seedUser("Alice", "alice@example.com", "Password@123")
	fromAcc := s.seedAccountWithBalance(alice, "THB", "500.00")

	w := s.POST("/api/v1/transactions/transfer", map[string]any{
		"fromAccountNumber": fromAcc.AccountNumber,
		"toAccountNumber":   "9999999999",
		"amount":            "100.00",
	}, s.tokenFor(alice))

	s.Equal(http.StatusNotFound, w.Code)
}

// --- Transaction list after operations ---

func (s *TransactionSuite) TestListTransactions_AfterDeposit() {
	user := s.seedUser("Alice", "alice@example.com", "Password@123")
	acc := s.seedAccount(user, "THB")

	s.POST("/api/v1/transactions/deposit", map[string]any{
		"accountNumber": acc.AccountNumber,
		"amount":        "100.00",
	}, s.tokenFor(user))

	w := s.GET(fmt.Sprintf("/api/v1/accounts/%s/transactions", acc.AccountNumber), s.tokenFor(user))
	s.Equal(http.StatusOK, w.Code)
	var body dtores.TransactionListResponse
	s.decodeBody(w, &body)
	s.Len(body.Transactions, 1)
	s.Equal(int64(1), body.Total)
}

func (s *TransactionSuite) TestListTransactions_AfterTransfer() {
	alice := s.seedUser("Alice", "alice@example.com", "Password@123")
	bob := s.seedUser("Bob", "bob@example.com", "Password@123")
	fromAcc := s.seedAccountWithBalance(alice, "THB", "500.00")
	toAcc := s.seedAccount(bob, "THB")

	s.POST("/api/v1/transactions/transfer", map[string]any{
		"fromAccountNumber": fromAcc.AccountNumber,
		"toAccountNumber":   toAcc.AccountNumber,
		"amount":            "200.00",
	}, s.tokenFor(alice))

	w := s.GET(fmt.Sprintf("/api/v1/accounts/%s/transactions", fromAcc.AccountNumber), s.tokenFor(alice))
	s.Equal(http.StatusOK, w.Code)
	var body dtores.TransactionListResponse
	s.decodeBody(w, &body)
	s.Equal(int64(1), body.Total)
}
