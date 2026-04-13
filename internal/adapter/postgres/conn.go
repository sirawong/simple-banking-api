package adapterdb

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/sirawong/simple-banking-api/internal/config"
	"github.com/sirawong/simple-banking-api/pkg/logger"
)

type DB struct {
	*gorm.DB
}

// @wire:set(name=AdapterSet)
func ProvideDB(cfg *config.Config) (*DB, func(), error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	})
	if err != nil {
		return nil, func() {}, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, func() {}, err
	}
	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, func() {}, fmt.Errorf("postgres: ping failed: %w", err)
	}

	cleanup := func() {
		sqlDB, err = db.DB()
		if err != nil {
			logger.Warn("failed to get sql.DB for cleanup", "error", err)
			return
		}
		if err = sqlDB.Close(); err != nil {
			logger.Warn("failed to close database connection", "error", err)
		}
	}

	return &DB{db}, cleanup, nil
}
