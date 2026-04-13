package testdi

import (
	"github.com/gin-gonic/gin"

	"github.com/sirawong/simple-banking-api/pkg/jwt"
)

type AppTestContainer struct {
	Router     *gin.Engine
	JWTManager jwt.JWTManager
}
