package auth

// import "go.mongodb.org/mongo-driver/bson/primitive"

// User defines the structure for a user
type User struct {
	Id              string `json:"_id" binding:"required" bson:"_id"`
	Name            string `json:"name" binding:"required"`
	Phone           string `json:"phone" binding:"required"`
	CountryCode     string `json:"country_code" binding:"required" bson:"country_code"`
	ProfilePicUrl   string `json:"profile_picture_url" bson:"profile_picture_url"`
	StatusMessage   string `json:"status_message" binding:"required" bson:"status_message"`
	LastSeen        string `json:"last_seen" bson:"last_seen"`
	IsOnline        int    `json:"is_online" bson:"is_online"`
	CreatedAt       string `json:"created_at" biniding:"required" bson:"created_at"`
	StatusUpdatedAt string `json:"updated_at" bson:"updated_at"`
}

type CappedUser struct {
	Name        string `json:"name" binding:"required"`
	Phone       string `json:"phone" binding:"required"`
	Password    string `json:"password" binding:"required"`
	CountryCode string `json:"country_code" binding:"required" bson:"country_code"`
}
