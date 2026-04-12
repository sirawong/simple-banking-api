package migrate

import (
	adapterdb "github.com/sirawong/simple-banking-api/internal/adapter/postgres"
	"github.com/sirawong/simple-banking-api/internal/repository/db/model"
)

type App struct {
	db *adapterdb.DB
}

// @wire:set(name=Migrate)
func ProvideMigrate(db *adapterdb.DB) *App {
	return &App{db: db}
}

func (a *App) Run() error {
	return a.db.AutoMigrate(
		&model.User{},
		&model.Account{},
		&model.Transaction{},
		&model.RefreshToken{},
	)
}
