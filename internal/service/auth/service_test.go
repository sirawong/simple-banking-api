package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"

	"github.com/sirawong/simple-banking-api/internal/config"
	"github.com/sirawong/simple-banking-api/internal/domain/entity"
	"github.com/sirawong/simple-banking-api/internal/errs"
	dbmock "github.com/sirawong/simple-banking-api/internal/repository/db/mock"
	pkgerrs "github.com/sirawong/simple-banking-api/pkg/errs"
	pkgjwt "github.com/sirawong/simple-banking-api/pkg/jwt"
)

type AuthServiceSuite struct {
	suite.Suite
	ur  *dbmock.MockUserRepository
	tr  *dbmock.MockTokenRepository
	svc Service
}

func (s *AuthServiceSuite) SetupTest() {
	s.ur = dbmock.NewMockUserRepository(s.T())
	s.tr = dbmock.NewMockTokenRepository(s.T())

	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:          "test-secret-key",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 168 * time.Hour,
		},
	}
	jwtManager := pkgjwt.ProvideManager(cfg)
	s.svc = ProvideService(cfg, jwtManager, s.ur, s.tr)
}

func TestAuthServiceSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceSuite))
}

func (s *AuthServiceSuite) TestRegister_Success() {
	created := &entity.User{ID: uuid.New(), Name: "Alice", Email: "alice@example.com"}
	s.ur.EXPECT().Create(mock.Anything, mock.AnythingOfType("*entity.User")).Return(created, nil)

	user, err := s.svc.Register(context.Background(), "Alice", "alice@example.com", "Password@123")
	s.NoError(err)
	s.Equal(created.ID, user.ID)
	s.Equal("Alice", user.Name)
}

func (s *AuthServiceSuite) TestRegister_DuplicateEmail() {
	s.ur.EXPECT().Create(mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil, errs.ErrDuplicateUser)

	_, err := s.svc.Register(context.Background(), "Alice", "alice@example.com", "Password@123")
	s.ErrorIs(err, errs.ErrDuplicateUser)
}

func (s *AuthServiceSuite) TestLogin_Success() {
	userID := uuid.New()
	hash, err := bcrypt.GenerateFromPassword([]byte("Password@123"), bcrypt.MinCost)
	s.Require().NoError(err)

	s.ur.EXPECT().FindByEmail(mock.Anything, "alice@example.com").Return(&entity.User{
		ID:           userID,
		Email:        "alice@example.com",
		PasswordHash: string(hash),
	}, nil)
	s.tr.EXPECT().Create(mock.Anything, mock.AnythingOfType("*entity.RefreshToken")).Return(&entity.RefreshToken{}, nil)

	pair, err := s.svc.Login(context.Background(), "alice@example.com", "Password@123")
	s.NoError(err)
	s.NotEmpty(pair.AccessToken)
	s.NotEmpty(pair.RefreshToken)
}

func (s *AuthServiceSuite) TestLogin_UserNotFound() {
	s.ur.EXPECT().FindByEmail(mock.Anything, "nobody@example.com").Return(nil, assert.AnError)

	_, err := s.svc.Login(context.Background(), "nobody@example.com", "Password@123")
	s.ErrorIs(err, errs.ErrInvalidPassword)
}

func (s *AuthServiceSuite) TestLogin_WrongPassword() {
	userID := uuid.New()
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	s.Require().NoError(err)

	s.ur.EXPECT().FindByEmail(mock.Anything, "alice@example.com").Return(&entity.User{
		ID:           userID,
		Email:        "alice@example.com",
		PasswordHash: string(hash),
	}, nil)

	_, err = s.svc.Login(context.Background(), "alice@example.com", "wrong-password")
	s.ErrorIs(err, errs.ErrInvalidPassword)
}

func (s *AuthServiceSuite) TestRefreshToken_Success() {
	userID := uuid.New()
	rawToken := "some-opaque-refresh-token"

	s.tr.EXPECT().FindByToken(mock.Anything, rawToken).Return(&entity.RefreshToken{
		ID:     uuid.New(),
		UserID: userID,
		Token:  rawToken,
	}, nil)
	s.ur.EXPECT().FindByID(mock.Anything, userID.String()).Return(&entity.User{
		ID:    userID,
		Email: "alice@example.com",
	}, nil)
	s.tr.EXPECT().DeleteByToken(mock.Anything, rawToken).Return(nil)
	s.tr.EXPECT().Create(mock.Anything, mock.AnythingOfType("*entity.RefreshToken")).Return(&entity.RefreshToken{}, nil)

	pair, err := s.svc.RefreshToken(context.Background(), rawToken)
	s.NoError(err)
	s.NotEmpty(pair.AccessToken)
	s.NotEmpty(pair.RefreshToken)
}

func (s *AuthServiceSuite) TestRefreshToken_InvalidToken() {
	s.tr.EXPECT().FindByToken(mock.Anything, "bad-token").Return(nil, assert.AnError)

	_, err := s.svc.RefreshToken(context.Background(), "bad-token")
	s.ErrorIs(err, pkgerrs.ErrInvalidToken)
}

func (s *AuthServiceSuite) TestRefreshToken_UserNotFound() {
	userID := uuid.New()
	rawToken := "some-opaque-refresh-token"

	s.tr.EXPECT().FindByToken(mock.Anything, rawToken).Return(&entity.RefreshToken{
		ID:     uuid.New(),
		UserID: userID,
		Token:  rawToken,
	}, nil)
	s.ur.EXPECT().FindByID(mock.Anything, userID.String()).Return(nil, assert.AnError)

	_, err := s.svc.RefreshToken(context.Background(), rawToken)
	s.ErrorIs(err, pkgerrs.ErrInvalidToken)
}
