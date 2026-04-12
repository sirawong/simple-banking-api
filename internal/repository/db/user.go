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
	if err := r.db.WithContext(ctx).Where("deleted_at IS NULL").First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrUserNotFound
		}
		return nil, err
	}
	return user.ToEntity(), nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("deleted_at IS NULL").First(&user, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrUserNotFound
		}
		return nil, err
	}
	return user.ToEntity(), nil
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	m := model.FromEntityUser(user)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		if isDuplicateError(err) {
			return errs.ErrDuplicateUser
		}
		return err
	}
	return nil
}
