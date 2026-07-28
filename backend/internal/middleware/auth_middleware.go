package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	jwtPkg "github.com/iamtbay/tyr-fintech/pkg/jwt"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := c.Cookie("accessToken")
		if err != nil {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				accessToken = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if accessToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization token required"})
			return
		}
		token, err := jwtPkg.VerifyToken(accessToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("userID", token.UserID)
		c.Next()
	}
}
