package db

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	adapterdb "github.com/sirawong/simple-banking-api/internal/adapter/postgres"
	"github.com/sirawong/simple-banking-api/internal/domain/entity"
	"github.com/sirawong/simple-banking-api/internal/repository/db/model"
	pkgerrs "github.com/sirawong/simple-banking-api/pkg/errs"
)

type tokenRepository struct {
	db *adapterdb.DB
}

// @wire:set(name=RepositorySet)
func ProvideTokenRepository(db *adapterdb.DB) TokenRepository {
	return &tokenRepository{db: db}
}

func (r *tokenRepository) Create(ctx context.Context, token *entity.RefreshToken) error {
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}
	m := model.FromEntityRefreshToken(token)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *tokenRepository) FindByToken(ctx context.Context, token string) (*entity.RefreshToken, error) {
	var m model.RefreshToken
	err := r.db.WithContext(ctx).Where("token = ? AND expires_at > ?", token, time.Now()).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrs.ErrInvalidToken
		}
		return nil, err
	}
	return m.ToEntity(), nil
}

func (r *tokenRepository) DeleteByToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("token = ?", token).Delete(&model.RefreshToken{}).Error
}

func (r *tokenRepository) DeleteByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.RefreshToken{}).Error
}
