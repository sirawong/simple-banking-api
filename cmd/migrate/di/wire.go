//go:build wireinject

package di

import (
	"github.com/google/wire"

	"github.com/sirawong/simple-banking-api/internal/di"
	"github.com/sirawong/simple-banking-api/internal/repository/db/migrate"
	"github.com/sirawong/simple-banking-api/pkg/logger"
)

func InitializeMigrate(log *logger.Logger) (*migrate.App, func(), error) {
	wire.Build(
		di.InternalSet,
		di.AdapterSet,
		di.Migrate,
	)
	return nil, nil, nil
}
