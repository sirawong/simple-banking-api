package testutil

import (
	"context"
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	adapterdb "github.com/sirawong/simple-banking-api/internal/adapter/postgres"
	"github.com/sirawong/simple-banking-api/internal/repository/db/model"
)

const defaultTestDSN = "postgres://postgres:postgres@localhost:5433/banking_test?sslmode=disable"

type TestDB struct {
	DB *adapterdb.DB
}

func StartTestDB(ctx context.Context) (*TestDB, error) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = defaultTestDSN
	}

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
