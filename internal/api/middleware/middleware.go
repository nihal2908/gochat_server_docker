package middleware

import (
	"gochat_server/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware checks for a valid JWT token in the request
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Missing Authorization token"})
			c.Abort()
			return
		}

		claims, err := utils.ValidateJWT(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid Authorization token"})
			c.Abort()
			return
		}

		// Pass the user claims to the request context
		c.Set("user", claims)
		c.Next()
	}
}
