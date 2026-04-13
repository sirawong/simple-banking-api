package utils

import (
	"github.com/gin-gonic/gin"

	dtores "github.com/sirawong/simple-banking-api/internal/handler/dto/response"
	"github.com/sirawong/simple-banking-api/internal/handler/middleware"
	"github.com/sirawong/simple-banking-api/pkg/errs"
)

func MustGetAuthUser(c *gin.Context) (*middleware.AuthUser, bool) {
	authUser, ok := middleware.GetAuthUser(c.Request.Context())
	if !ok {
		dtores.HandleResponse(c, nil, errs.ErrUnauthorized)
		return nil, false
	}
	return authUser, true
}
