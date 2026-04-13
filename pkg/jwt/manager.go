package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/sirawong/simple-banking-api/internal/config"
	"github.com/sirawong/simple-banking-api/pkg/errs"
)

// jwtManager handles signing and parsing of JWT access tokens.
type jwtManager struct {
	secret []byte
	ttl    time.Duration
}

type JWTManager interface {
	TTL() time.Duration
	Generate(userID, email string) (string, error)
	Parse(tokenStr string) (*Claims, error)
}

// ProvideJWTManager constructs a manager from application config.
func ProvideJWTManager(cfg *config.Config) JWTManager {
	return &jwtManager{secret: []byte(cfg.JWT.Secret), ttl: cfg.JWT.AccessTokenTTL}
}

// TTL returns the access token lifetime.
func (m *jwtManager) TTL() time.Duration {
	return m.ttl
}

// Generate creates a signed JWT for the given user.
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

// Parse validates a JWT string and returns the embedded claims.
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
