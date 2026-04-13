package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/sirawong/simple-banking-api/internal/handler/dto/response"
	pkgerrs "github.com/sirawong/simple-banking-api/pkg/errs"
	pkgjwt "github.com/sirawong/simple-banking-api/pkg/jwt"
)

type authContextKey struct{}

type AuthUser struct {
	ID    string
	Email string
}

// GetAuthUser retrieves the authenticated user from the request context.
func GetAuthUser(ctx context.Context) (*AuthUser, bool) {
	u, ok := ctx.Value(authContextKey{}).(*AuthUser)
	return u, ok
}

// JWTAuth returns a middleware that validates Bearer tokens and injects auth claims into context.
func JWTAuth(manager pkgjwt.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.HandleResponse(c, nil, pkgerrs.ErrUnauthorized)
			c.Abort()
			return
		}

		claims, err := manager.Parse(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			response.HandleResponse(c, nil, pkgerrs.ErrInvalidToken)
			c.Abort()
			return
		}

		ctx := context.WithValue(c.Request.Context(), authContextKey{}, &AuthUser{
			ID:    claims.UserID,
			Email: claims.Email,
		})
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
