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

func (r *tokenRepository) Create(ctx context.Context, token *entity.RefreshToken) (*entity.RefreshToken, error) {
	tk := model.FromEntityRefreshToken(token)
	if tk.ID == uuid.Nil {
		tk.ID = uuid.New()
	}
	if err := r.db.WithContext(ctx).Create(tk).Error; err != nil {
		return nil, err
	}
	return tk.ToEntity(), nil
}

func (r *tokenRepository) FindByToken(ctx context.Context, token string) (*entity.RefreshToken, error) {
	var rt model.RefreshToken
	err := r.db.WithContext(ctx).
		Where("token = ? AND expires_at > ?", token, time.Now()).
		First(&rt).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrs.ErrInvalidToken
		}
		return nil, err
	}
	return rt.ToEntity(), nil
}

func (r *tokenRepository) DeleteByToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("token = ?", token).Delete(&model.RefreshToken{}).Error
}
