package group

import (
	"context"
	"net/http"
	"time"

	"gochat_server/internal/db"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Create a new group
func CreateGroup(c *gin.Context) {
	var newGroup Group
	if err := c.ShouldBindJSON(&newGroup); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	newGroup.ID = primitive.NewObjectID().Hex()
	newGroup.CreatedAt = time.Now().Format(time.RFC3339)
	newGroup.UpdatedAt = newGroup.CreatedAt
	// Iterate over existing members and update their JoinedAt
	for i := range newGroup.Members {
		newGroup.Members[i].JoinedAt = newGroup.CreatedAt // Set all members' JoinedAt to current time
	}
	// Add the creator separately
	newGroup.Members = append(newGroup.Members, GroupMember{
		UserID:   newGroup.CreatedBy,
		IsAdmin:  true,
		JoinedAt: newGroup.CreatedAt,
	})
	GroupCollection := db.GetCollection("groups")
	_, err := GroupCollection.InsertOne(context.TODO(), newGroup)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Group created successfully", "group": newGroup})
}

// Join an existing group
func JoinGroup(c *gin.Context) {
	groupID := c.Param("id")
	var member GroupMember

	if err := c.ShouldBindJSON(&member); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	GroupCollection := db.GetCollection("groups")
	update := bson.M{"$push": bson.M{"members": member}, "$set": bson.M{"updated_at": time.Now().Format(time.RFC3339)}}
	_, err := GroupCollection.UpdateOne(context.TODO(), bson.M{"_id": groupID}, update)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User added to group"})
}

// Delete a group (only by creator)
func DeleteGroup(c *gin.Context) {
	groupID := c.Param("id")
	userID := c.Query("user_id") // Assuming user ID is passed as a query param
	GroupCollection := db.GetCollection("groups")
	// Check if the user is the creator
	var group Group
	err := GroupCollection.FindOne(context.TODO(), bson.M{"_id": groupID}).Decode(&group)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	if group.CreatedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the creator can delete the group"})
		return
	}

	_, err = GroupCollection.DeleteOne(context.TODO(), bson.M{"_id": groupID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group deleted successfully"})
}

// Update group details (Only Admins can update)
func UpdateGroup(c *gin.Context) {
	groupID := c.Param("id")
	userID := c.Query("user_id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	GroupCollection := db.GetCollection("groups")
	// Check if the user is an admin
	var group Group
	err := GroupCollection.FindOne(context.TODO(), bson.M{"_id": groupID}).Decode(&group)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	isAdmin := false
	for _, member := range group.Members {
		if member.UserID == userID && member.IsAdmin {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can update group details"})
		return
	}

	updates["updated_at"] = time.Now().Format(time.RFC3339)
	_, err = GroupCollection.UpdateOne(context.TODO(), bson.M{"_id": groupID}, bson.M{"$set": updates})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group updated successfully"})
}

// Leave a group
func LeaveGroup(c *gin.Context) {
	groupID := c.Param("id")
	userID := c.Query("user_id")
	GroupCollection := db.GetCollection("groups")
	update := bson.M{"$pull": bson.M{"members": bson.M{"user_id": userID}}, "$set": bson.M{"updated_at": time.Now().Format(time.RFC3339)}}
	_, err := GroupCollection.UpdateOne(context.TODO(), bson.M{"_id": groupID}, update)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to leave group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User left the group"})
}

// Get Group Data
func GetGroupData(c *gin.Context) {
	groupID := c.Param("id")
	GroupCollection := db.GetCollection("groups")
	var group Group
	err := GroupCollection.FindOne(context.TODO(), bson.M{"_id": groupID}).Decode(&group)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"group": group})
}
