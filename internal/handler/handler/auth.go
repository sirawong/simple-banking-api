package handler

import (
	"github.com/gin-gonic/gin"

	dtoreq "github.com/sirawong/simple-banking-api/internal/handler/dto/request"
	dtores "github.com/sirawong/simple-banking-api/internal/handler/dto/response"
	authsvc "github.com/sirawong/simple-banking-api/internal/service/auth"
)

type AuthHandler struct {
	authSvc authsvc.Service
}

// @wire:set(name=HandlerSet)
func ProvideAuthHandler(authSvc authsvc.Service) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// Register godoc
// @Summary      Register
// @Description  Create a new user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dtoreq.RegisterRequest  true  "Register request"
// @Success      201   {object}  dtores.UserResponse
// @Failure      400   {object}  errs.AppError
// @Failure      409   {object}  errs.AppError
// @Router       /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dtoreq.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dtores.HandleCreatedResponse(c, nil, err)
		return
	}

	user, err := h.authSvc.Register(c.Request.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		dtores.HandleCreatedResponse(c, nil, err)
		return
	}
	dtores.HandleCreatedResponse(c, dtores.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, nil)
}

// Login godoc
// @Summary      Login
// @Description  Authenticate with email and password, returns JWT token pair
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dtoreq.LoginRequest  true  "Login request"
// @Success      200   {object}  dtores.TokenPairResponse
// @Failure      400   {object}  errs.AppError
// @Failure      401   {object}  errs.AppError
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dtoreq.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dtores.HandleResponse(c, nil, err)
		return
	}

	pair, err := h.authSvc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		dtores.HandleResponse(c, nil, err)
		return
	}
	dtores.HandleResponse(c, dtores.TokenPairResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    pair.ExpiresIn,
	}, nil)
}

// RefreshToken godoc
// @Summary      Refresh token
// @Description  Exchange a valid refresh token for a new token pair
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dtoreq.RefreshTokenRequest  true  "Refresh token request"
// @Success      200   {object}  dtores.TokenPairResponse
// @Failure      401   {object}  errs.AppError
// @Router       /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dtoreq.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dtores.HandleResponse(c, nil, err)
		return
	}

	pair, err := h.authSvc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		dtores.HandleResponse(c, nil, err)
		return
	}
	dtores.HandleResponse(c, dtores.TokenPairResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    pair.ExpiresIn,
	}, nil)
}
