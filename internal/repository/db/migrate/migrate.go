package migrate

import (
	"gorm.io/gorm"

	"github.com/sirawong/simple-banking-api/internal/domain/entity"
)

type App struct {
	db *gorm.DB
}

// @wire:set(name=Migrate)
func ProvideMigrate(db *gorm.DB) *App {
	return &App{db: db}
}

func (a *App) Run() error {
	return a.db.AutoMigrate(
		&entity.User{},
		&entity.Account{},
		&entity.Transaction{},
	)
}
