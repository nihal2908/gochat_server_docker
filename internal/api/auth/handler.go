package auth

import (
	"context"
	"fmt"
	"gochat_server/internal/db"
	"gochat_server/internal/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// SigninRequest represents the structure of the login request payload
type LoginRequest struct {
	Phone       string `json:"phone" binding:"required"`
	Password    string `json:"password" binding:"required"`
	CountryCode string `json:"required"`
}

// SigninHandler handles user authentication
func LoginHandler(c *gin.Context) {
	var request LoginRequest

	// Parse and validate the incoming JSON payload
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if the user exists in the database
	collection := db.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var foundUser CappedUser
	err := collection.FindOne(ctx, bson.M{"phone": request.Phone}).Decode(&foundUser)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid phone or password"})
		return
	}

	// Compare the provided password with the hashed password in the database
	err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(request.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid phone or password"})
		return
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user User
	err = collection.FindOne(ctx, bson.M{"phone": request.Phone}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Error fetching details."})
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(foundUser.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}

	// Respond with the generated token
	c.JSON(http.StatusOK, gin.H{
		"message": "Signin successful",
		"token":   token,
		"user":    user,
	})
}

func RegisterHandler(c *gin.Context) {
	var newUser CappedUser

	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if userExists(newUser.Phone) {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this phone number already exists"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}

	user := bson.M{
		"name":                newUser.Name,
		"phone":               newUser.Phone,
		"country_code":        newUser.CountryCode,
		"is_online":           0,
		"created_at":          time.Now().Format(time.RFC3339),
		"updated_at":          time.Now().Format(time.RFC3339),
		"profile_picture_url": "",
		"last_seen":           "",
		"status_message":      "Hey there! Lets connect.",
		"password":            string(hashedPassword),
	}

	// Insert the new user into the database
	collection := db.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(newUser.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating JWT token"})
		return
	}

	// Respond with a success message and JWT token
	c.JSON(http.StatusOK, gin.H{
		"message": "User registered successfully",
		"token":   token,
		"user":    newUser,
	})
}

// userExists checks if a user with the given email already exists in the database
func userExists(phone string) bool {
	collection := db.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"phone": phone}
	var result bson.M
	err := collection.FindOne(ctx, filter).Decode(&result)
	return err == nil // If no error, it means the user exists
}

func GetUserDataHandler(c *gin.Context) {
	userId := c.Query("userId")
	if userId == "" {
		fmt.Printf("userId not found in query.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	collection := db.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		fmt.Printf("Invalid ObjectId string: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid userId"})
		return
	}

	filter := bson.M{"_id": objectID}
	projection := bson.M{
		"_id":                 1,
		"name":                1,
		"phone":               1,
		"country_code":        1,
		"profile_picture_url": 1,
		"last_seen":           1,
		"is_online":           1,
		"status_message":      1,
		"created_at":          1,
		"updated_at":          1,
	}

	var user User
	err = collection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&user)

	if err != nil {
		fmt.Printf("No user found with userId: %s", userId)
		c.JSON(http.StatusNotFound, bson.M{"error": "no user found"})
		return
	}

	c.JSON(http.StatusOK, user)
}
