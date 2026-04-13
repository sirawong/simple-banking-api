package testutil

import (
	"context"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	adapterdb "github.com/sirawong/simple-banking-api/internal/adapter/postgres"
	"github.com/sirawong/simple-banking-api/internal/config"
	"github.com/sirawong/simple-banking-api/internal/repository/db/model"
)

type TestDB struct {
	DB *adapterdb.DB
}

func StartTestDB(cfg *config.Config) (*TestDB, error) {

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("open gorm: %w", err)
	}

	db := &adapterdb.DB{DB: gormDB}

	if err := db.AutoMigrate(
		&model.User{},
		&model.Account{},
		&model.Transaction{},
		&model.RefreshToken{},
	); err != nil {
		return nil, fmt.Errorf("automigrate: %w", err)
	}

	return &TestDB{DB: db}, nil
}

func (t *TestDB) Close() error {
	sqlDB, err := t.DB.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (t *TestDB) TruncateAll(ctx context.Context) error {
	return t.DB.WithContext(ctx).Exec(
		`TRUNCATE TABLE refresh_tokens, transactions, accounts, users RESTART IDENTITY CASCADE`,
	).Error
}
