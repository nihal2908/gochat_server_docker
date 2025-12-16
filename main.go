package main

import (
	"gochat_server/config"
	"gochat_server/internal/db"
	"gochat_server/pkg/server"
	"log"
)

func main() {
	config.LoadConfig()
	db.ConnectDB()

	// Create the Gin router
	r := server.NewRouter()

	// Start the server on port 8080
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Error starting the server: %v", err)
	}
}
