package server

import (
	"github.com/gin-gonic/gin"
)

// New initializes a new Gin router
func New() *gin.Engine {
	return gin.Default()
}
