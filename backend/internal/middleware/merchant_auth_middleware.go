package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/iamtbay/tyr-fintech/internal/models"
	"github.com/iamtbay/tyr-fintech/internal/repos"
)

func MerchantAuthRequired(merchantRepo repos.MerchantRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: API key required in Authorization header",
			})
			return
		}
		rawKey := strings.TrimPrefix(authHeader, "Bearer ")
		if rawKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: Empty API key provided",
			})
			return
		}

		hasher := sha256.New()
		hasher.Write([]byte(rawKey))
		keyHash := hex.EncodeToString(hasher.Sum(nil))

		merchant, err := merchantRepo.GetMerchantByAPIKeyHash(c.Request.Context(), keyHash)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: Invalid or inactive API key",
			})
			return
		}
		c.Set("merchant", merchant)
		c.Set("merchantID", merchant.ID)
		c.Next()
	}

}

func GetAuthenticatedMerchant(c *gin.Context) (*models.Merchant, bool) {
	val, ok := c.Get("merchant")
	if !ok {
		return nil, false
	}
	merchant, ok := val.(*models.Merchant)
	return merchant, ok
}
