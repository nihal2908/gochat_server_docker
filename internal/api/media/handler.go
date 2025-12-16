package media

import (
	"bytes"
	"fmt"
	"gochat_server/internal/db"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
)

// UploadImage uploads an image to MongoDB GridFS and returns the image URL
func UploadImage(c *gin.Context) {
	// Get the file from the request
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get image file"})
		return
	}
	defer file.Close()

	// Get GridFS bucket
	bucket, err := gridfs.NewBucket(db.GetDB())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize GridFS"})
		return
	}

	// Read file into a buffer
	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// Upload the file to GridFS
	uploadStream, err := bucket.OpenUploadStream(header.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}
	defer uploadStream.Close()

	_, err = uploadStream.Write(buf.Bytes())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write image to GridFS"})
		return
	}

	// Get the image ID
	imageID := uploadStream.FileID.(primitive.ObjectID).Hex()

	// Generate a URL (Assuming an endpoint will serve images)
	imageURL := fmt.Sprintf("/media/image/%s", imageID)

	// Return the URL in response
	c.JSON(http.StatusOK, gin.H{"image_url": imageURL})
}

// ServeImage retrieves an image from MongoDB GridFS
func ServeImage(c *gin.Context) {
	imageID := c.Param("id")

	// Convert ID to ObjectID
	objID, err := primitive.ObjectIDFromHex(imageID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image ID"})
		return
	}

	// Get GridFS bucket
	bucket, err := gridfs.NewBucket(db.GetDB())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize GridFS"})
		return
	}

	// Open a download stream
	var buf bytes.Buffer
	_, err = bucket.DownloadToStream(objID, &buf)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	// Serve the image
	c.Header("Content-Type", "image/jpeg")
	c.Writer.Write(buf.Bytes())
}
