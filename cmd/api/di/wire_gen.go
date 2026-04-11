//go:build !wireinject

package di

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	adapterdb "github.com/sirawong/simple-banking-api/internal/adapter/postgres"
	adapterredis "github.com/sirawong/simple-banking-api/internal/adapter/redis"
	"github.com/sirawong/simple-banking-api/internal/config"
	internalDi "github.com/sirawong/simple-banking-api/internal/di"
	internalhandler "github.com/sirawong/simple-banking-api/internal/handler"
	"github.com/sirawong/simple-banking-api/internal/handler/handler"
	redisrepo "github.com/sirawong/simple-banking-api/internal/repository/cache"
	"github.com/sirawong/simple-banking-api/internal/service/account"
	"github.com/sirawong/simple-banking-api/internal/service/transaction"
	"github.com/sirawong/simple-banking-api/pkg/logger"
)

var _ = wire.NewSet
var _ = internalDi.InternalSet

func InitializeApp(cfg *config.Config) (*gin.Engine, error) {
	db, err := adapterdb.ProvideDB(cfg)
	if err != nil {
		return nil, err
	}
	rdb := adapterredis.ProvideRedisClient(cfg)

	accountRepo := db.ProvideAccountRepository(db)
	txRepo := db.ProvideTransactionRepository(db)
	cacheRepo := redisrepo.ProvideCacheRepository(rdb)

	_ = db.ProvideUserRepository(db)

	accountSvc := account.ProvideService(accountRepo, cacheRepo)
	txSvc := transaction.ProvideService(db, accountRepo, txRepo, cacheRepo)

	accountHandler := handler.ProvideAccountHandler(accountSvc, txSvc)
	txHandler := handler.ProvideTransactionHandler(txSvc)

	log := provideLogger()
	engine := internalhandler.ProvideRouter(accountHandler, txHandler, log)
	return engine, nil
}

func provideLogger() *logger.Logger {
	return logger.ProvideGlobalLogger()
}
