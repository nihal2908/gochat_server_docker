package file

import (
	"context"
	"image"
	"io"
	"net/http"
	"os"

	"gochat_server/internal/db"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file upload"})
		return
	}
	defer file.Close()

	// Read first 512 bytes for MIME detection
	head := make([]byte, 512)
	n, _ := file.Read(head)
	mimeType := detectMimeType(head[:n])
	mediaType := classifyMediaType(mimeType)

	// Reset file reader
	file.Seek(0, io.SeekStart)

	bucket, err := gridfs.NewBucket(db.GetDB())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize GridFS"})
		return
	}

	uploadOpts := options.GridFSUpload().
		SetMetadata(bson.M{
			"mime_type":  mimeType,
			"media_type": mediaType,
			"file_name":  header.Filename,
			"size":       header.Size,
		})

	uploadStream, err := bucket.OpenUploadStream(header.Filename, uploadOpts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload stream"})
		return
	}
	defer uploadStream.Close()

	// Buffer + optional image processing
	buffer := make([]byte, 1024*1024)

	var (
		blurHash string
		width    int
		height   int
	)

	if mediaType == "image" {
		img, _, err := image.Decode(file)
		if err == nil {
			blurHash, width, height, _ = generateBlurHash(img)
		}
		file.Seek(0, io.SeekStart)
	}

	if mediaType == "video" {
		tmp, err := os.CreateTemp("", "upload-*.mp4")
		if err == nil {
			io.Copy(tmp, file)
			tmp.Close()

			blurHash, width, height, _ = blurHashFromVideo(tmp.Name())
			os.Remove(tmp.Name())

			file.Seek(0, io.SeekStart)
		}
	}


	for {
		n, err := file.Read(buffer)
		if n > 0 {
			_, writeErr := uploadStream.Write(buffer[:n])
			if writeErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write file"})
				return
			}
		}
		if err == io.EOF {
			break
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading file"})
			return
		}
	}

	fileID := uploadStream.FileID.(primitive.ObjectID)

	// Return FULL media descriptor
	c.JSON(http.StatusOK, gin.H{
		"media_id":   fileID.Hex(),
		"media_type": mediaType,
		"mime_type":  mimeType,
		"size":       header.Size,
		"file_name":  header.Filename,
		"width":      width,
		"height":     height,
		"blur_hash":  blurHash,
	})
}


func DownloadFile(c *gin.Context) {
	fileIDHex := c.Param("file_id")
	fileID, err := primitive.ObjectIDFromHex(fileIDHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	bucket, err := gridfs.NewBucket(db.GetDB())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize GridFS"})
		return
	}

	var fileDoc bson.M
	err = bucket.GetFilesCollection().
		FindOne(context.TODO(), bson.M{"_id": fileID}).
		Decode(&fileDoc)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	mimeType, _ := fileDoc["metadata"].(bson.M)["mime_type"].(string)
	fileName, _ := fileDoc["metadata"].(bson.M)["file_name"].(string)

	downloadStream, err := bucket.OpenDownloadStream(fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open download stream"})
		return
	}
	defer downloadStream.Close()

	c.Writer.Header().Set("Content-Type", mimeType)
	c.Writer.Header().Set("Content-Disposition", "inline; filename=\""+fileName+"\"")

	buffer := make([]byte, 1024*1024)
	for {
		n, err := downloadStream.Read(buffer)
		if n > 0 {
			c.Writer.Write(buffer[:n])
			c.Writer.Flush()
		}
		if err == io.EOF {
			break
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading file"})
			return
		}
	}
}
