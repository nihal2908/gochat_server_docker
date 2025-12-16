package server

import (
	"gochat_server/internal/api/auth"
	"gochat_server/internal/api/chat"
	"gochat_server/internal/api/contacts"
	"gochat_server/internal/api/fcm"
	"gochat_server/internal/api/group"
	"gochat_server/internal/api/media"
	"gochat_server/internal/api/file"
	"gochat_server/internal/api/websocket"

	"github.com/gin-gonic/gin"
)

// NewRouter initializes the Gin router and defines routes
func NewRouter() *gin.Engine {
	r := gin.Default()

	// API v1 group
	api := r.Group("/api")
	{
		// Auth routes
		api.POST("/login", auth.LoginHandler)
		api.POST("/register", auth.RegisterHandler)

		api.GET("/ws", websocket.WebSocketHandler)

		api.POST("/match-contacts", contacts.MatchContactsHandler)

		api.POST("/fcm/store-fcm-token", fcm.StoreFCMToken)
		api.POST("/fcm/unset-fcm-token", fcm.UnsetFCMToken)

		api.GET("/userdata", auth.GetUserDataHandler)

		api.GET("/chats", chat.GetChatsHandler)

		api.POST("/groups/create-group", group.CreateGroup)
		api.DELETE("/groups/delete-group", group.DeleteGroup)
		api.POST("/groups/join-group/:id", group.JoinGroup)
		api.DELETE("/groups/leave-group", group.LeaveGroup)
		api.PUT("/groups/update-group", group.UpdateGroup)
		api.GET("/groups/get-group-data/:id", group.GetGroupData)

		api.POST("/media/upload-image", media.UploadImage)
		api.GET("/media/image/:id", media.ServeImage)

		api.POST("/file/upload", file.UploadFile)
		api.GET("file/download/:file_id", file.DownloadFile)
	}

	return r
}
