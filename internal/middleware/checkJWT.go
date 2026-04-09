package middleware

import (
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/token"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func CheckJWT(j *token.JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		const prefix = "Bearer "
		if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		jwt := strings.TrimSpace(auth[len(prefix):])

		userID, err := j.ParseJWT(jwt)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set("userID", userID)
		c.Next()
	}
}
