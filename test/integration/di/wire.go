//go:build wireinject

package testdi

import (
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"

	adapterdb "github.com/sirawong/simple-banking-api/internal/adapter/postgres"
	"github.com/sirawong/simple-banking-api/internal/config"
	"github.com/sirawong/simple-banking-api/internal/di"
	pkgdi "github.com/sirawong/simple-banking-api/pkg/di"
	"github.com/sirawong/simple-banking-api/pkg/logger"
)

func InitializeTestContainer(
	cfg *config.Config,
	db *adapterdb.DB,
	rdb *redis.Client,
	log *logger.Logger,
) (*AppTestContainer, error) {
	wire.Build(
		pkgdi.PkgSet,
		di.RepositorySet,
		di.ServiceSet,
		di.HandlerSet,
		wire.Struct(new(AppTestContainer), "*"),
	)
	return nil, nil
}
