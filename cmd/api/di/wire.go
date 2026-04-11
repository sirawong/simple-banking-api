//go:build wireinject

package di

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"github.com/sirawong/simple-banking-api/internal/di"
	"github.com/sirawong/simple-banking-api/internal/server"
	"github.com/sirawong/simple-banking-api/pkg/logger"
)

func InitializeApp(log *logger.Logger) (*server.App, func(), error) {
	wire.Build(
		di.InternalSet,
		di.AdapterSet,
		di.RepositorySet,
		di.ServiceSet,
		di.HandlerSet,
		wire.Bind(new(http.Handler), new(*gin.Engine)),
	)
	return nil, nil, nil
}
