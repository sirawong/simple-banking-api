package db

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	adapterdb "github.com/sirawong/simple-banking-api/internal/adapter/postgres"
	"github.com/sirawong/simple-banking-api/internal/domain/entity"
	"github.com/sirawong/simple-banking-api/internal/errs"
	"github.com/sirawong/simple-banking-api/internal/repository/db/model"
	pkgerrs "github.com/sirawong/simple-banking-api/pkg/errs"
)

type userRepository struct {
	db *adapterdb.DB
}

// @wire:set(name=RepositorySet)
func ProvideUserRepository(db *adapterdb.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Scopes(notDeleted).
		Where("id = ?", id).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrUserNotFound
		}
		return nil, err
	}
	return user.ToEntity(), nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Scopes(notDeleted).
		Where("email = ?", email).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrUserNotFound
		}
		return nil, err
	}
	return user.ToEntity(), nil
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	u := model.FromEntityUser(user)
	if u == nil {
		return nil, pkgerrs.ErrBadRequest
	}
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	if err := r.db.WithContext(ctx).Create(u).Error; err != nil {
		if isDuplicateError(err) {
			return nil, errs.ErrDuplicateUser
		}
		return nil, err
	}
	return u.ToEntity(), nil
}
