package jwt

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	UserID string `json:"sub"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}
