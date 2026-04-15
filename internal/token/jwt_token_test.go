package token

import (
	"testing"
	"time"

	"github.com/MaxDrattcev/loyalty_system_service.git/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func newTestJWT(secret string, expiresMin int32) JWT {
	cfg := &config.Config{
		JWTToken: config.JWTTokenConfig{
			Secret:    secret,
			ExpiresAt: expiresMin,
		},
	}
	return NewJWT(cfg)
}

func TestJWT_BuildAndParse_Success(t *testing.T) {
	j := newTestJWT("test-secret", 60)

	tokenString, err := j.BuildJWTString(42, "max")
	require.NoError(t, err)
	require.NotEmpty(t, tokenString)

	userID, err := j.ParseJWT(tokenString)
	require.NoError(t, err)
	require.Equal(t, int64(42), userID)
}

func TestJWT_ParseJWT_TableDriven(t *testing.T) {
	goodJWT := newTestJWT("good-secret", 60)
	otherJWT := newTestJWT("other-secret", 60)

	validToken, err := goodJWT.BuildJWTString(7, "user")
	require.NoError(t, err)

	claimsEmptyUser := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		UserID: 0,
		Login:  "user",
	}
	tokenEmptyUser := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsEmptyUser)
	emptyUserTokenString, err := tokenEmptyUser.SignedString([]byte("good-secret"))
	require.NoError(t, err)

	claimsExpired := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
		UserID: 9,
		Login:  "expired-user",
	}
	tokenExpired := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsExpired)
	expiredTokenString, err := tokenExpired.SignedString([]byte("good-secret"))
	require.NoError(t, err)

	tests := []struct {
		name        string
		jwtInst     JWT
		tokenString string
		wantUserID  int64
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid token",
			jwtInst:     goodJWT,
			tokenString: validToken,
			wantUserID:  7,
			wantErr:     false,
		},
		{
			name:        "invalid token format",
			jwtInst:     goodJWT,
			tokenString: "not-a-jwt",
			wantErr:     true,
			errContains: "parse jwt",
		},
		{
			name:        "wrong secret",
			jwtInst:     otherJWT,
			tokenString: validToken,
			wantErr:     true,
			errContains: "parse jwt",
		},
		{
			name:        "expired token",
			jwtInst:     goodJWT,
			tokenString: expiredTokenString,
			wantErr:     true,
			errContains: "parse jwt",
		},
		{
			name:        "empty user id in claims",
			jwtInst:     goodJWT,
			tokenString: emptyUserTokenString,
			wantErr:     true,
			errContains: "invalid token: empty user id",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			gotUserID, err := tt.jwtInst.ParseJWT(tt.tokenString)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					require.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantUserID, gotUserID)
		})
	}
}
