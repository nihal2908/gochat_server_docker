package contacts

import (
	"context"
	"gochat_server/internal/api/auth"
	"gochat_server/internal/db"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Request struct {
	Contacts []string `json:"contacts"`
}

type Response struct {
	MatchedUsers []auth.User `json:"matched_users"`
}

func MatchContactsHandler(c *gin.Context) {
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// MongoDB query to match hashed phone numbers
	filter := bson.M{"phone": bson.M{"$in": req.Contacts}}
	projection := bson.M{
		"_id": 1, 
		"name": 1, 
		"phone": 1,
		"country_code": 1, 
		"profile_picture_url":1, 
		"last_seen":1,
		"is_online":1,
		"status_message":1,
		"created_at":1,
		"updated_at":1,
	}

	userCollection := db.GetCollection("users")
	cursor, err := userCollection.Find(context.TODO(), filter, options.Find().SetProjection(projection))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}
	defer cursor.Close(context.TODO())

	// Process results
	var matchedUsers []auth.User
	for cursor.Next(context.TODO()) {
		var user auth.User
		if err := cursor.Decode(&user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding user data"})
			return
		}
		matchedUsers = append(matchedUsers, user)
	}

	// Check for cursor errors
	if err := cursor.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cursor error"})
		return
	}

	// Respond with matched users
	c.JSON(http.StatusOK, Response{MatchedUsers: matchedUsers})
}