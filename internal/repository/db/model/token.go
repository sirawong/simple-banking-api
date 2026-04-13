package model

import (
	"time"

	"github.com/google/uuid"

	"github.com/sirawong/simple-banking-api/internal/domain/entity"
)

type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;index;not null"`
	Token     string    `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time

	User User `gorm:"foreignKey:UserID"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

func (t *RefreshToken) ToEntity() *entity.RefreshToken {
	if t == nil {
		return nil
	}
	return &entity.RefreshToken{
		ID:        t.ID,
		UserID:    t.UserID,
		Token:     t.Token,
		ExpiresAt: t.ExpiresAt,
		CreatedAt: t.CreatedAt,
	}
}

func FromEntityRefreshToken(t *entity.RefreshToken) *RefreshToken {
	if t == nil {
		return nil
	}
	return &RefreshToken{
		ID:        t.ID,
		UserID:    t.UserID,
		Token:     t.Token,
		ExpiresAt: t.ExpiresAt,
		CreatedAt: t.CreatedAt,
	}
}
