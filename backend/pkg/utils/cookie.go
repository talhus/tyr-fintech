package utils

import (
	"os"

	"github.com/gin-gonic/gin"
)

func SetAuthCookie(c *gin.Context, token string) {
	isSecure := os.Getenv("ENV") == "production" // true if production
	c.SetCookie("accessToken", token, 3600*24, "/", "", isSecure, true)
}

func ClearAuthCookie(c *gin.Context) {
	isSecure := os.Getenv("ENV") == "production"
	c.SetCookie("accessToken", "", -1, "/", "", isSecure, true)
}
