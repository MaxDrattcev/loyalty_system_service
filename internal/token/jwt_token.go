// Package token provides JWT generation and parsing utilities for authentication.
package token

import (
	"fmt"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

// JWT provides methods to build and parse JWT tokens.
type JWT struct {
	cfg *config.Config
}

// NewJWT creates JWT helper with provided application config.
func NewJWT(cfg *config.Config) JWT {
	return JWT{
		cfg: cfg,
	}
}

// Claims represents JWT payload with application-specific user fields.
type Claims struct {
	jwt.RegisteredClaims
	UserID int64  `json:"user_id"`
	Login  string `json:"login"`
}

// BuildJWTString builds and signs JWT token for user identity.
func (j *JWT) BuildJWTString(userID int64, login string) (string, error) {
	now := time.Now()
	claims := &Claims{

		RegisteredClaims: jwt.RegisteredClaims{

			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(j.cfg.JWTToken.ExpiresAt) * time.Minute)),
		},
		UserID: userID,
		Login:  login,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.cfg.JWTToken.Secret))
	if err != nil {
		return "", fmt.Errorf("error signing token: %v", err)
	}
	return tokenString, nil
}

// ParseJWT validates token and returns user ID from claims.
func (j *JWT) ParseJWT(tokenString string) (int64, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(j.cfg.JWTToken.Secret), nil
	})
	if err != nil {
		return 0, fmt.Errorf("parse jwt: %w", err)
	}
	if claims.UserID == 0 {
		return 0, fmt.Errorf("invalid token: empty user id")
	}
	return claims.UserID, nil
}
