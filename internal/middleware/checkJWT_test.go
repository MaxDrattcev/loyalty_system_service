package middleware

import (
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/config"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/token"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		JWTToken: config.JWTTokenConfig{
			Secret:    "test-secret",
			ExpiresAt: 60,
		},
	}
	j := token.NewJWT(cfg)
	makeRouter := func() *gin.Engine {
		r := gin.New()
		r.Use(CheckJWT(&j))
		r.GET("/protected", func(c *gin.Context) {
			userID, ok := c.Get("userID")
			if !ok {
				c.Status(http.StatusInternalServerError)
				return
			}
			c.JSON(http.StatusOK, gin.H{"user_id": userID})
		})
		return r
	}
	t.Run("missing authorization header -> 401", func(t *testing.T) {
		r := makeRouter()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("invalid token -> 401", func(t *testing.T) {
		r := makeRouter()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("valid token -> 200 and userID in context", func(t *testing.T) {
		r := makeRouter()
		jwtString, err := j.BuildJWTString(42, "max")
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+jwtString)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), `"user_id":42`)
	})
}
