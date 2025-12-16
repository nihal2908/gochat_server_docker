package fcm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"golang.org/x/oauth2/google"
)

var (
	accessToken     string
	tokenExpiryTime time.Time
	mutex           sync.Mutex
)

func SendFCMWakeSignal(receiverID string) {
	deviceToken, err := getDeviceTokenForUser(receiverID)
	if err != nil {
		fmt.Printf("Failed to get token for %s: %v\n", receiverID, err)
		return
	}

	err = sendNotification(deviceToken)
	if err != nil {
		fmt.Printf("Failed to send wake signal for %s: %v\n", receiverID, err)
	}
}

// Function to get or refresh the OAuth 2.0 access token
func getAccessToken() (string, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if time.Now().Before(tokenExpiryTime) {
		return accessToken, nil
	}

	// Try getting path from env var, fallback to relative path
	credPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_FILE_PATH")
	if credPath == "" {
		return "", fmt.Errorf("google application credentials path not found")
	}

	credBytes, err := os.ReadFile(credPath)
	if err != nil {
		return "", fmt.Errorf("failed to read credentials file (%s): %v", credPath, err)
	}

	conf, err := google.JWTConfigFromJSON(credBytes, "https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		return "", fmt.Errorf("failed to parse credentials: %v", err)
	}

	token, err := conf.TokenSource(context.Background()).Token()
	if err != nil {
		return "", fmt.Errorf("failed to get token: %v", err)
	}

	accessToken = token.AccessToken
	tokenExpiryTime = token.Expiry

	return accessToken, nil
}

func sendNotification(deviceToken string) error {
	token, err := getAccessToken()
	if err != nil {
		return fmt.Errorf("failed to get access token: %v", err)
	}

	fcmURL := "https://fcm.googleapis.com/v1/projects/whatsapp-clone-77193/messages:send"

	// Convert messageContent struct into map[string]string with nested struct handling
	// msgMap := structToMap(messageContent)

	fcmMessage := map[string]interface{}{
		"message": map[string]interface{}{
			"token": deviceToken,
			"data":  map[string]string{"signal": "wake"},
			// // Optional: use notification payload for visible banner alerts
			// "notification": map[string]string{
			// 	"title": "New Message",
			// 	"body":  "You received a new message",
			// },
		},
	}

	messageBytes, err := json.Marshal(fcmMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	req, err := http.NewRequest("POST", fcmURL, bytes.NewReader(messageBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to send FCM message, status: %v, body: %s", resp.StatusCode, body)
	}

	fmt.Println("Message sent successfully!")
	return nil
}