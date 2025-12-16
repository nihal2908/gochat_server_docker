package chat

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetChatsHandler retrieves a list of chats
func GetChatsHandler(c *gin.Context) {
	// Logic to fetch user chats
	c.JSON(http.StatusOK, gin.H{"message": "Fetched chats successfully"})
}