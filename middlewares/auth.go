package middlewares

import (
	"chronote/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTAuthMiddlewares() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Auth Header
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Without Token, and you're unauthorized!",
			})
			c.Abort()
			return
		}

		// Get Valid Token Part
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Token format is invalid!",
			})
			c.Abort()
			return
		}
		tokenString := parts[1]

		// Verify Token Effectiveness and Type
		claims, err := utils.Parsetoken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Token is invalid or outdated!",
			})
			c.Abort()
			return
		}
		if claims.TokenType != "access" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Need Use Aceess Token to Authorize!",
			})
			c.Abort()
			return
		}

		// Process Next Part With User Info
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Name)
		c.Next()
	}
}
