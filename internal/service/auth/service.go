package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/sirawong/simple-banking-api/internal/config"
	"github.com/sirawong/simple-banking-api/internal/domain/entity"
	"github.com/sirawong/simple-banking-api/internal/errs"
	dbrepo "github.com/sirawong/simple-banking-api/internal/repository/db"
	pkgerrs "github.com/sirawong/simple-banking-api/pkg/errs"
	pkgjwt "github.com/sirawong/simple-banking-api/pkg/jwt"
)

type Service interface {
	Register(ctx context.Context, name, email, password string) (*entity.User, error)
	Login(ctx context.Context, email, password string) (*entity.TokenPair, error)
	RefreshToken(ctx context.Context, refreshToken string) (*entity.TokenPair, error)
}

type service struct {
	jwtManager pkgjwt.Manager
	refreshTTL time.Duration
	userRepo   dbrepo.UserRepository
	tokenRepo  dbrepo.TokenRepository
}

// @wire:set(name=ServiceSet)
func ProvideService(cfg *config.Config, jwtManager pkgjwt.Manager, userRepo dbrepo.UserRepository, tokenRepo dbrepo.TokenRepository) Service {
	return &service{
		jwtManager: jwtManager,
		refreshTTL: cfg.JWT.RefreshTokenTTL,
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
	}
}

func (s *service) Register(ctx context.Context, name, email, password string) (*entity.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, pkgerrs.ErrInternal.WithError(err)
	}

	user := &entity.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *service) Login(ctx context.Context, email, password string) (*entity.TokenPair, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, errs.ErrInvalidPassword
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errs.ErrInvalidPassword
	}

	return s.issueTokenPair(ctx, user)
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*entity.TokenPair, error) {
	stored, err := s.tokenRepo.FindByToken(ctx, refreshToken)
	if err != nil {
		return nil, pkgerrs.ErrInvalidToken
	}

	user, err := s.userRepo.FindByID(ctx, stored.UserID.String())
	if err != nil {
		return nil, pkgerrs.ErrInvalidToken
	}

	if err := s.tokenRepo.DeleteByToken(ctx, refreshToken); err != nil {
		return nil, pkgerrs.ErrInternal.WithError(err)
	}

	return s.issueTokenPair(ctx, user)
}

func (s *service) issueTokenPair(ctx context.Context, user *entity.User) (*entity.TokenPair, error) {
	accessToken, err := s.jwtManager.Generate(user.ID.String(), user.Email)
	if err != nil {
		return nil, pkgerrs.ErrInternal.WithError(err)
	}

	rawRefresh, err := generateOpaqueToken()
	if err != nil {
		return nil, pkgerrs.ErrInternal.WithError(err)
	}

	rt := &entity.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     rawRefresh,
		ExpiresAt: time.Now().Add(s.refreshTTL),
	}
	if err := s.tokenRepo.Create(ctx, rt); err != nil {
		return nil, pkgerrs.ErrInternal.WithError(err)
	}

	return &entity.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    int64(s.jwtManager.TTL().Seconds()),
	}, nil
}

// generateOpaqueToken generates a cryptographically random token for use as a refresh token.
func generateOpaqueToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
