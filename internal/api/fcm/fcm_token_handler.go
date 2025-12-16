package fcm

import (
	"context"
	"fmt"
	"gochat_server/internal/db"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type StoreFCMRequest struct {
	UserId   string `json:"_id"`
	FCMToken string `json:"fcm_token"`
}

func StoreFCMToken(c *gin.Context) {
	var request StoreFCMRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		fmt.Printf("Invalid request payload.")
		c.JSON(http.StatusBadRequest, bson.M{"error": "bad request"})
		return
	}

	objectId, err := primitive.ObjectIDFromHex(request.UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, bson.M{"error": "error creating object id."})
		return
	}

	collection := db.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = collection.UpdateByID(ctx, objectId, bson.M{"$set": bson.M{"fcm_token": request.FCMToken}})
	if err != nil {
		fmt.Printf("No user found with userId: %s, %s", request.UserId, err.Error())
		c.JSON(http.StatusNotFound, bson.M{"error": "no user found"})
		return
	}

	c.JSON(http.StatusOK, bson.M{"success": "success"})
}

func UnsetFCMToken(c *gin.Context) {
	var request StoreFCMRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		fmt.Printf("Invalid request payload.")
		c.JSON(http.StatusBadRequest, bson.M{"error": "bad request"})
		return
	}

	objectId, err := primitive.ObjectIDFromHex(request.UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, bson.M{"error": "error creating object id."})
		return
	}

	collection := db.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use "$unset" to remove the "fcm_token" field
	result, err := collection.UpdateByID(ctx, objectId, bson.M{"$unset": bson.M{"fcm_token": ""}})
	if err != nil {
		fmt.Printf("Error unsetting FCM token for userId: %s, %s", request.UserId, err.Error())
		c.JSON(http.StatusInternalServerError, bson.M{"error": "error unsetting FCM token"})
		return
	}

	// Check if the update matched a document
	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, bson.M{"error": "no user found"})
		return
	}

	c.JSON(http.StatusOK, bson.M{"success": "fcm token removed successfully"})
}

func getDeviceTokenForUser(userId string) (string, error) {
	collection := db.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		fmt.Printf("Invalid ObjectId string: %s", err.Error())
	}

	filter := bson.M{"_id": objectID}
	var result struct {
		FCMToken string `bson:"fcm_token"`
	}

	err = collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", fmt.Errorf("no FCM token found for user ID: %s", userId)
		}
		return "", fmt.Errorf("failed to retrieve FCM token: %v", err)
	}

	return result.FCMToken, nil
}