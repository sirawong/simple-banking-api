//go:build wireinject

package testdi

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"

	adapterdb "github.com/sirawong/simple-banking-api/internal/adapter/postgres"
	"github.com/sirawong/simple-banking-api/internal/config"
	"github.com/sirawong/simple-banking-api/internal/di"
	pkgdi "github.com/sirawong/simple-banking-api/pkg/di"
	"github.com/sirawong/simple-banking-api/pkg/logger"
)

// InitializeRouter wires the full handler stack for integration tests.
// db and rdb are provided externally; config is loaded from ENV_FILE env var.
func InitializeRouter(
	db *adapterdb.DB,
	rdb *redis.Client,
	log *logger.Logger,
) (*gin.Engine, error) {
	wire.Build(
		config.ProvideConfig,
		pkgdi.PkgSet,
		di.RepositorySet,
		di.ServiceSet,
		di.HandlerSet,
	)
	return nil, nil
}
