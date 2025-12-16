package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"gochat_server/internal/api/fcm"
	"gochat_server/internal/db"
	"gochat_server/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	onlineUsers = make(map[string]*websocket.Conn)
	onlineUsersMutex sync.RWMutex

	messageHandlers = map[string]func(incmsg IncomingMessage){
		"message":           handleMessageType,
		"ack_read":          handleReadAck,
		"ack_sent":          handleSentAck,
		"ack_delivered":     handleDeliveredAck,
		"edit_message":      handleEditMessage,
		"delete_message":    handleDeleteMessage,
		"webrtc_offer":      handleWebRTCOffer,
		"webrtc_answer":     handleWebRTCAnswer,
		"webrtc_candidate":  handleICECandidate,
		"webrtc_delivered":  handleWebRTCDelivered,
		"webrtc_hangup":     handleWebRTCHangup,
		"webrtc_decline":    handleWebRTCDecline,
	}
)

func WebSocketHandler(c *gin.Context) {
	userId := c.Query("userId")
	if userId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing userId in query parameters"})
		return
	}

	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Error upgrading connection:", err)
		return
	}

	if !markOnline(userId, conn) {
		fmt.Println("Error marking user online")
		conn.Close()
		return
	}

	// After WebSocket connection is established
	go deliverOfflineMessages(userId, conn)

	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Println("Error closing WebSocket:", err)
		}

		onlineUsersMutex.Lock()
		delete(onlineUsers, userId)
		onlineUsersMutex.Unlock()

		fmt.Println("WebSocket connection closed for user:", userId)
	}()


	fmt.Println("WebSocket connection established for user:", userId)

	// Handle WebSocket communication
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}

		fmt.Println("Message received from user", userId, ":", string(message))

		var incmsg IncomingMessage
		err = json.Unmarshal(message, &incmsg)
		if err != nil {
			fmt.Println("Invalid message payload1" + err.Error())
			continue
		}

		if handler, ok := messageHandlers[incmsg.Type]; ok {
			handler(incmsg)
		} else {
			fmt.Println("Unknown message type:", incmsg.Type)
		}
	}
}

func handleMessageType(incmsg IncomingMessage) {
	var message Message
	if err := utils.BindData(incmsg.Data, &message); err != nil {
		fmt.Println("Invalid Message Payload" + err.Error())
		return
	}

	message.ServerTS = time.Now().Format(time.RFC3339)
	message.Status = "sent"

	var sentAck SentAcknowledgment
	sentAck.MessageId = message.Id
	sentAck.SenderId = message.SenderId
	sentAck.ReceiverId = message.ReceiverId
	sentAck.Timestamp = message.Timestamp
	sentAck.ServerTS = message.ServerTS
	sentAck.ChatId = message.ChatId
	sentAck.GroupId = message.GroupId

	sendJsonMessage(
		message.SenderId, 
		map[string]interface{}{
			"type": "ack_sent",
			"data": sentAck,
		},
	)

	sendJsonMessage(message.ReceiverId, incmsg)
}

func handleEditMessage(incmsg IncomingMessage) {
	var message Message
	if err := utils.BindData(incmsg.Data, &message); err != nil {
		fmt.Println("Error binding message:", err)
		return
	}

	var sentAck SentAcknowledgment
	sentAck.MessageId = message.Id
	sentAck.SenderId = message.SenderId
	sentAck.ReceiverId = message.ReceiverId
	sentAck.ServerTS = message.ServerTS
	sentAck.Timestamp = message.Timestamp
	sentAck.ChatId = message.ChatId
	sentAck.GroupId = message.GroupId

	sendJsonMessage(message.SenderId, map[string]interface{}{
		"type": "ack_sent",
		"data": sentAck,
	})

	message.Edited = 1

	sendJsonMessage(message.ReceiverId, map[string]interface{}{
		"type": "edit_message",
		"data": incmsg.Data,
	})
}

func handleDeleteMessage(incmsg IncomingMessage) {
	var deleteMessage DeletedForEveryoneMessage
	if err := utils.BindData(incmsg.Data, &deleteMessage); err != nil {
		fmt.Println("Error binding message:", err)
		return
	}

	var sentAck SentAcknowledgment
	sentAck.MessageId = deleteMessage.Id
	sentAck.SenderId = deleteMessage.SenderId
	sentAck.ReceiverId = deleteMessage.ReceiverId
	sentAck.Timestamp = deleteMessage.Timestamp
	sentAck.ServerTS = deleteMessage.ServerTS
	sentAck.ChatId = deleteMessage.ChatId
	sentAck.GroupId = deleteMessage.GroupId

	sendJsonMessage(deleteMessage.SenderId, map[string]interface{}{
		"type": "ack_sent",
		"data": sentAck,
	})

	sendJsonMessage(deleteMessage.ReceiverId, map[string]interface{}{
		"type": "delete_message",
		"data": incmsg.Data,
	})
}

func handleReadAck(incmsg IncomingMessage) {
	var ackData ReadAcknowledgment
	if err := utils.BindData(incmsg.Data, &ackData); err != nil {
		fmt.Println("Error binding read acknowledgment:", err)
		return
	}
	// fmt.Printf("Read acknowledgment received: %+v\n", ackData)
	sendJsonMessage(ackData.SenderId, incmsg)
}

func handleSentAck(incmsg IncomingMessage) {
	var ackData SentAcknowledgment
	if err := utils.BindData(incmsg.Data, &ackData); err != nil {
		fmt.Println("Error binding sent acknowledgment:", err)
		return
	}
	// fmt.Printf("Sent acknowledgment received: %+v\n", ackData)
	sendJsonMessage(ackData.ReceiverId, incmsg)
}

func handleDeliveredAck(incmsg IncomingMessage) {
	var ackData DeliveredAcknowledgment
	if err := utils.BindData(incmsg.Data, &ackData); err != nil {
		fmt.Println("Error binding delivered acknowledgment:", err)
		return
	}

	// Send acknowledgment to sender
	sendJsonMessage(ackData.SenderId, incmsg)

	// Remove message from MongoDB
	messagesCollection := db.GetCollection("offline_messages")
	_, err := messagesCollection.DeleteOne(context.TODO(), bson.M{
		"data.message_id": ackData.MessageId,
	})
	if err != nil {
		fmt.Println("Failed to delete delivered message:", err)
	}
}


func handleWebRTCOffer(incmsg IncomingMessage) {
	sendJsonMessage(utils.GetReceiverId(incmsg.Data), incmsg)
}

func handleWebRTCAnswer(incmsg IncomingMessage) {
	sendJsonMessage(utils.GetReceiverId(incmsg.Data), incmsg)
}

func handleWebRTCDelivered(incmsg IncomingMessage) {
	sendJsonMessage(utils.GetReceiverId(incmsg.Data), incmsg)
}

func handleICECandidate(incmsg IncomingMessage) {
	sendJsonMessage(utils.GetReceiverId(incmsg.Data), incmsg)
}

func handleWebRTCHangup(incmsg IncomingMessage) {
	sendJsonMessage(utils.GetReceiverId(incmsg.Data), incmsg)
}

func handleWebRTCDecline(incmsg IncomingMessage) {
	sendJsonMessage(utils.GetReceiverId(incmsg.Data), incmsg)
}


// markOnline marks the user as online by storing their WebSocket connection
func markOnline(userId string, conn *websocket.Conn) bool {
	if userId == "" {
		return false
	}
	onlineUsersMutex.Lock()
	defer onlineUsersMutex.Unlock()

	onlineUsers[userId] = conn
	return true
}

func getOnlineUser(userId string) (*websocket.Conn, bool) {
	onlineUsersMutex.RLock()
	defer onlineUsersMutex.RUnlock()
	conn, exists := onlineUsers[userId]
	return conn, exists
}

func sendJsonMessage(receiverId string, data interface{}) error {
	recConn, exists := getOnlineUser(receiverId)
	if exists {
		// Receiver is online — send immediately
		return recConn.WriteJSON(data)
	}

	// Receiver is offline — send FCM wake signal only
	storeOfflineMessage(receiverId, data)

	return nil
}

func storeOfflineMessage(receiverId string, data interface{}) {
	go fcm.SendFCMWakeSignal(receiverId)

	// Assert that data is a map (it should be!)
	msgMap := make(map[string]interface{})

	// Add receiver_id and timestamp directly into the message map
	msgMap["receiver_id"] = receiverId
	msgMap["timestamp"] = time.Now().Format(time.RFC3339)
	msgMap["message"] = data

	_, err := db.GetCollection("offline_messages").InsertOne(context.TODO(), msgMap)
	if err != nil {
		fmt.Println("Failed to store message in DB:", err)
	} else {
		fmt.Printf("Stored message for offline user %s\n", receiverId)
	}
}

func deliverOfflineMessages(userId string, conn *websocket.Conn) {
	messagesCollection := db.GetCollection("offline_messages")

	filter := bson.M{"receiver_id": userId}
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}})

	cursor, err := messagesCollection.Find(context.TODO(), filter, opts)
	if err != nil {
		fmt.Println("Error fetching offline messages:", err)
		return
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var msg bson.M
		if err := cursor.Decode(&msg); err != nil {
			fmt.Println("Error decoding message:", err)
			continue
		}
		
		if err := conn.WriteJSON(msg["message"]); err != nil {
			fmt.Printf("Failed to deliver message to %s: %v\n", userId, err)
		}
	}
}
