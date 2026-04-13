package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/sirawong/simple-banking-api/internal/config"
	"github.com/sirawong/simple-banking-api/pkg/errs"
)

type jwtManager struct {
	secret []byte
	ttl    time.Duration
}

type JWTManager interface {
	TTL() time.Duration
	Generate(userID, email string) (string, error)
	Parse(tokenStr string) (*Claims, error)
}

func ProvideJWTManager(cfg *config.Config) JWTManager {
	return &jwtManager{secret: []byte(cfg.JWT.Secret), ttl: cfg.JWT.AccessTokenTTL}
}

func (m *jwtManager) TTL() time.Duration {
	return m.ttl
}

func (m *jwtManager) Generate(userID, email string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *jwtManager) Parse(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errs.ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, errs.ErrInvalidToken
	}
	return claims, nil
}
