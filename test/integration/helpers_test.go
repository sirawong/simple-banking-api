package integration_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/sirawong/simple-banking-api/internal/domain/entity"
	"github.com/sirawong/simple-banking-api/internal/repository/db/model"
	"github.com/sirawong/simple-banking-api/pkg/errs"
)

func (s *BaseSuite) tokenFor(user *entity.User) string {
	tok, err := s.jwtManager.Generate(user.ID.String(), user.Email)
	s.Require().NoError(err)
	return tok
}

type errBody struct {
	Message string `json:"error_message"`
}

func (s *BaseSuite) requireErrMessage(w *httptest.ResponseRecorder, want *errs.AppError) {
	var body errBody
	s.decodeBody(w, &body)
	s.Equal(want.Message, body.Message)
}

// seedUser inserts a user directly into the DB and returns the entity.
func (s *BaseSuite) seedUser(name, email, password string) *entity.User {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	s.Require().NoError(err)

	m := &model.User{
		ID:           uuid.New(),
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
	}
	s.Require().NoError(s.db.Create(m).Error)
	return m.ToEntity()
}

// seedAccount inserts an account for the given user and returns the entity.
func (s *BaseSuite) seedAccount(user *entity.User, currency string) *entity.Account {
	m := &model.Account{
		ID:            uuid.New(),
		UserID:        user.ID,
		AccountNumber: nextAccountNumber(),
		Currency:      currency,
	}
	s.Require().NoError(s.db.Create(m).Error)
	return m.ToDomain()
}

// seedAccountWithBalance inserts an account with a pre-set balance.
func (s *BaseSuite) seedAccountWithBalance(user *entity.User, currency string, balance string) *entity.Account {
	m := &model.Account{
		ID:            uuid.New(),
		UserID:        user.ID,
		AccountNumber: nextAccountNumber(),
		Currency:      currency,
	}
	s.Require().NoError(s.db.Create(m).Error)
	s.Require().NoError(
		s.db.Model(m).Update("balance", balance).Error,
	)
	err := m.Balance.Scan(balance)
	s.Require().NoError(err)
	return m.ToDomain()
}

// nextAccountNumber returns a unique 10-digit account number for each call.
var accountSeq atomic.Int64

func nextAccountNumber() string {
	n := accountSeq.Add(1)
	return fmt.Sprintf("%010d", 1000000000+n)
}

func (s *BaseSuite) do(method, path string, body any, token string) *httptest.ResponseRecorder {
	var buf *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		s.Require().NoError(err)
		buf = bytes.NewBuffer(b)
	} else {
		buf = &bytes.Buffer{}
	}

	req := httptest.NewRequest(method, path, buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	return w
}

func (s *BaseSuite) GET(path, token string) *httptest.ResponseRecorder {
	return s.do(http.MethodGet, path, nil, token)
}

func (s *BaseSuite) POST(path string, body any, token string) *httptest.ResponseRecorder {
	return s.do(http.MethodPost, path, body, token)
}

func (s *BaseSuite) decodeBody(w *httptest.ResponseRecorder, dst any) {
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), dst))
}
