package model

import (
	"time"

	"github.com/google/uuid"

	"github.com/sirawong/simple-banking-api/internal/domain/entity"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name         string    `gorm:"not null"`
	Email        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time `gorm:"index"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) ToEntity() *entity.User {
	if u == nil {
		return nil
	}
	return &entity.User{
		ID:           u.ID,
		Name:         u.Name,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

func FromEntityUser(u *entity.User) *User {
	if u == nil {
		return nil
	}
	return &User{
		ID:           u.ID,
		Name:         u.Name,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}
