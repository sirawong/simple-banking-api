package handler

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/sirawong/simple-banking-api/docs"
	handlerhttp "github.com/sirawong/simple-banking-api/internal/handler/handler"
	"github.com/sirawong/simple-banking-api/internal/handler/middleware"
	pkgjwt "github.com/sirawong/simple-banking-api/pkg/jwt"
	"github.com/sirawong/simple-banking-api/pkg/logger"
)

// @wire:set(name=HandlerSet)
func ProvideRouter(
	jwtManager pkgjwt.Manager,
	authHandler *handlerhttp.AuthHandler,
	accountHandler *handlerhttp.AccountHandler,
	txHandler *handlerhttp.TransactionHandler,
	log *logger.Logger,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger(log))

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)

		protected := v1.Group("", middleware.JWTAuth(jwtManager))
		{
			accounts := protected.Group("/accounts")
			accounts.POST("", accountHandler.CreateAccount)
			accounts.GET("/:id", accountHandler.GetAccount)
			accounts.GET("/:id/transactions", accountHandler.ListTransactions)

			transactions := protected.Group("/transactions")
			transactions.POST("/deposit", txHandler.Deposit)
			transactions.POST("/withdraw", txHandler.Withdraw)
			transactions.POST("/transfer", txHandler.Transfer)
		}
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
