package utils

import (
	"github.com/gin-gonic/gin"

	dtores "github.com/sirawong/simple-banking-api/internal/handler/dto/response"
	"github.com/sirawong/simple-banking-api/internal/handler/middleware"
	"github.com/sirawong/simple-banking-api/pkg/errs"
)

// MustGetAuthUser extracts the authenticated user from the context.
// If not present, it writes a 401 response and returns false.
func MustGetAuthUser(c *gin.Context) (*middleware.AuthUser, bool) {
	authUser, ok := middleware.GetAuthUser(c.Request.Context())
	if !ok {
		dtores.HandleResponse(c, nil, errs.ErrUnauthorized)
		return nil, false
	}
	return authUser, true
}
