//go:build wireinject

package di

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	adapterdb "github.com/sirawong/simple-banking-api/internal/adapter/postgres"
	adapterredis "github.com/sirawong/simple-banking-api/internal/adapter/redis"
	"github.com/sirawong/simple-banking-api/internal/config"
	"github.com/sirawong/simple-banking-api/internal/di"
	"github.com/sirawong/simple-banking-api/pkg/logger"
)

func InitializeApp(cfg *config.Config) (*gin.Engine, error) {
	wire.Build(
		adapterdb.ProvideDB,
		adapterredis.ProvideRedisClient,
		provideLogger,
		di.RepositorySet,
		di.ServiceSet,
		di.HandlerSet,
	)
	return nil, nil
}

func provideLogger() *logger.Logger {
	return logger.ProvideGlobalLogger()
}
