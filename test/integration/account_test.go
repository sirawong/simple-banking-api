package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/sirawong/simple-banking-api/internal/errs"
	dtores "github.com/sirawong/simple-banking-api/internal/handler/dto/response"
)

type AccountSuite struct{ BaseSuite }

func TestAccountSuite(t *testing.T) { suite.Run(t, new(AccountSuite)) }

// --- Create Account ---

func (s *AccountSuite) TestCreateAccount_Success() {
	user := s.seedUser("Alice", "alice@example.com", "Password@123")
	token := s.tokenFor(user)

	w := s.POST("/api/v1/accounts", map[string]any{
		"currency": "THB",
	}, token)

	s.Equal(http.StatusCreated, w.Code)
	var body dtores.AccountResponse
	s.decodeBody(w, &body)
	s.Equal(user.ID, body.UserID)
	s.Equal("THB", body.Currency)
	s.NotEmpty(body.AccountNumber)
}

func (s *AccountSuite) TestCreateAccount_Unauthorized() {
	w := s.POST("/api/v1/accounts", map[string]any{"currency": "THB"}, "")
	s.Equal(http.StatusUnauthorized, w.Code)
}

func (s *AccountSuite) TestCreateAccount_MissingCurrency() {
	user := s.seedUser("Alice", "alice@example.com", "Password@123")
	w := s.POST("/api/v1/accounts", map[string]any{}, s.tokenFor(user))
	s.Equal(http.StatusBadRequest, w.Code)
}

// --- List Accounts ---

func (s *AccountSuite) TestListAccounts_ReturnOnlyOwnerAccounts() {
	alice := s.seedUser("Alice", "alice@example.com", "Password@123")
	bob := s.seedUser("Bob", "bob@example.com", "Password@123")

	s.seedAccount(alice, "THB")
	s.seedAccount(alice, "USD")
	s.seedAccount(bob, "THB")

	w := s.GET("/api/v1/accounts", s.tokenFor(alice))

	s.Equal(http.StatusOK, w.Code)
	var body []*dtores.AccountResponse
	s.decodeBody(w, &body)
	s.Len(body, 2)
	for _, acc := range body {
		s.Equal(alice.ID, acc.UserID)
	}
}

func (s *AccountSuite) TestListAccounts_EmptyWhenNoAccounts() {
	user := s.seedUser("Alice", "alice@example.com", "Password@123")
	w := s.GET("/api/v1/accounts", s.tokenFor(user))
	s.Equal(http.StatusOK, w.Code)
	var body []*dtores.AccountResponse
	s.decodeBody(w, &body)
	s.Empty(body)
}

// --- Get Account (Balance) ---

func (s *AccountSuite) TestGetAccount_Success() {
	user := s.seedUser("Alice", "alice@example.com", "Password@123")
	acc := s.seedAccountWithBalance(user, "THB", "500.00")

	w := s.GET(fmt.Sprintf("/api/v1/accounts/%s", acc.AccountNumber), s.tokenFor(user))

	s.Equal(http.StatusOK, w.Code)
	var body dtores.BalanceResponse
	s.decodeBody(w, &body)
	s.Equal(acc.AccountNumber, body.AccountNumber)
	s.True(body.Balance.IsPositive())
}

func (s *AccountSuite) TestGetAccount_NotFound() {
	user := s.seedUser("Alice", "alice@example.com", "Password@123")
	w := s.GET("/api/v1/accounts/9999999999", s.tokenFor(user))
	s.Equal(http.StatusNotFound, w.Code)
	s.requireErrMessage(w, errs.ErrAccountNotFound)
}

func (s *AccountSuite) TestGetAccount_ForbiddenOtherUser() {
	alice := s.seedUser("Alice", "alice@example.com", "Password@123")
	bob := s.seedUser("Bob", "bob@example.com", "Password@123")
	acc := s.seedAccount(alice, "THB")

	w := s.GET(fmt.Sprintf("/api/v1/accounts/%s", acc.AccountNumber), s.tokenFor(bob))
	s.Equal(http.StatusForbidden, w.Code)
	s.requireErrMessage(w, errs.ErrForbidden)
}

// --- List Transactions ---

func (s *AccountSuite) TestListTransactions_EmptyInitially() {
	user := s.seedUser("Alice", "alice@example.com", "Password@123")
	acc := s.seedAccount(user, "THB")

	w := s.GET(fmt.Sprintf("/api/v1/accounts/%s/transactions", acc.AccountNumber), s.tokenFor(user))

	s.Equal(http.StatusOK, w.Code)
	var body dtores.TransactionListResponse
	s.decodeBody(w, &body)
	s.Empty(body.Transactions)
	s.Equal(int64(0), body.Total)
}

func (s *AccountSuite) TestListTransactions_ForbiddenOtherUser() {
	alice := s.seedUser("Alice", "alice@example.com", "Password@123")
	bob := s.seedUser("Bob", "bob@example.com", "Password@123")
	acc := s.seedAccount(alice, "THB")

	w := s.GET(fmt.Sprintf("/api/v1/accounts/%s/transactions", acc.AccountNumber), s.tokenFor(bob))
	s.Equal(http.StatusForbidden, w.Code)
}
