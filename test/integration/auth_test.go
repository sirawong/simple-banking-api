package integration_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/sirawong/simple-banking-api/internal/errs"
	dtores "github.com/sirawong/simple-banking-api/internal/handler/dto/response"
)

type AuthSuite struct{ BaseSuite }

func TestAuthSuite(t *testing.T) { suite.Run(t, new(AuthSuite)) }

func (s *AuthSuite) TestRegister_Success() {
	w := s.POST("/api/v1/auth/register", map[string]any{
		"name":     "Alice",
		"email":    "alice@example.com",
		"password": "Password@123",
	}, "")

	s.Equal(http.StatusCreated, w.Code)
	var body dtores.UserResponse
	s.decodeBody(w, &body)
	s.Equal("alice@example.com", body.Email)
	s.Equal("Alice", body.Name)
	s.NotEmpty(body.ID)
}

func (s *AuthSuite) TestRegister_DuplicateEmail() {
	s.seedUser("Alice", "alice@example.com", "Password@123")

	w := s.POST("/api/v1/auth/register", map[string]any{
		"name":     "Alice2",
		"email":    "alice@example.com",
		"password": "Password@123",
	}, "")

	s.Equal(http.StatusConflict, w.Code)
	s.requireErrMessage(w, errs.ErrDuplicateUser)
}

func (s *AuthSuite) TestRegister_InvalidBody() {
	w := s.POST("/api/v1/auth/register", map[string]any{
		"email": "not-an-email",
	}, "")
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *AuthSuite) TestLogin_Success() {
	s.seedUser("Bob", "bob@example.com", "Password@123")

	w := s.POST("/api/v1/auth/login", map[string]any{
		"email":    "bob@example.com",
		"password": "Password@123",
	}, "")

	s.Equal(http.StatusOK, w.Code)
	var body dtores.TokenPairResponse
	s.decodeBody(w, &body)
	s.NotEmpty(body.AccessToken)
	s.NotEmpty(body.RefreshToken)
	s.Equal("Bearer", body.TokenType)
}

func (s *AuthSuite) TestLogin_WrongPassword() {
	s.seedUser("Bob", "bob@example.com", "Password@123")

	w := s.POST("/api/v1/auth/login", map[string]any{
		"email":    "bob@example.com",
		"password": "WrongPassword",
	}, "")

	s.Equal(http.StatusUnauthorized, w.Code)
	s.requireErrMessage(w, errs.ErrInvalidPassword)
}

func (s *AuthSuite) TestLogin_UserNotFound() {
	w := s.POST("/api/v1/auth/login", map[string]any{
		"email":    "nobody@example.com",
		"password": "Password@123",
	}, "")
	s.Equal(http.StatusUnauthorized, w.Code)
}

func (s *AuthSuite) TestRefreshToken_Success() {
	s.seedUser("Carol", "carol@example.com", "Password@123")
	loginW := s.POST("/api/v1/auth/login", map[string]any{
		"email": "carol@example.com", "password": "Password@123",
	}, "")
	var tokens dtores.TokenPairResponse
	s.decodeBody(loginW, &tokens)

	w := s.POST("/api/v1/auth/refresh", map[string]any{
		"refreshToken": tokens.RefreshToken,
	}, "")

	s.Equal(http.StatusOK, w.Code)
	var body dtores.TokenPairResponse
	s.decodeBody(w, &body)
	s.NotEmpty(body.AccessToken)
	s.NotEmpty(body.RefreshToken)
	// refresh token must rotate
	s.NotEqual(tokens.RefreshToken, body.RefreshToken)
}

func (s *AuthSuite) TestRefreshToken_InvalidToken() {
	w := s.POST("/api/v1/auth/refresh", map[string]any{
		"refreshToken": "invalid-token",
	}, "")
	s.Equal(http.StatusUnauthorized, w.Code)
}
