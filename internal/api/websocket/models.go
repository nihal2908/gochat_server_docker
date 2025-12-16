package websocket

type Message struct {
	Id                 string `json:"_id"`
	SenderId           string `json:"sender_id"`
	ReceiverId         string `json:"receiver_id"`
	Content            string `json:"content"`
	Timestamp          string `json:"timestamp"`
	ServerTS		   string `json:"server_ts"`
	ChatId             string `json:"chat_id"`
	GroupId            string `json:"group_id"`
	Type               string `json:"type"`
	Status             string `json:"status"`
	DeletedForEveryone int    `json:"deleted_for_everyone"`
	Edited             int    `json:"edited"`
}

type ReadAcknowledgment struct {
	SenderId   string `json:"sender_id"`
	ReceiverId string `json:"receiver_id"`
	ChatId     string `json:"chat_id"`
	GroupId    string `json:"group_id"`
	Timestamp  string `json:"timestamp"`
}

type DeliveredAcknowledgment struct {
	MessageId  string `json:"message_id"`
	SenderId   string `json:"sender_id"`
	ReceiverId string `json:"receiver_id"`
	ChatId     string `json:"chat_id"`
	GroupId    string `json:"group_id"`
	Timestamp  string `json:"timestamp"`
}

type SentAcknowledgment struct {
	MessageId  string `json:"message_id"`
	SenderId   string `json:"sender_id"`
	ReceiverId string `json:"receiver_id"`
	ChatId     string `json:"chat_id"`
	GroupId    string `json:"group_id"`
	Timestamp  string `json:"timestamp"`
	ServerTS   string `json:"server_ts"`
}

type DeletedForEveryoneMessage struct {
	Id         string `json:"_id"`
	SenderId   string `json:"sender_id"`
	ReceiverId string `json:"receiver_id"`
	ChatId     string `json:"chat_id"`
	GroupId    string `json:"group_id"`
	Timestamp  string `json:"timestamp"`
	ServerTS   string `json:"server_ts"`
}

type ErrorAcknowledgment struct {
	MessageId  string `json:"message_id"`
	SenderId   string `json:"sender_id"`
	ReceiverId string `json:"receiver_id"`
	ChatId     string `json:"chat_id"`
	GroupId    string `json:"group_id"`
	Timestamp  string `json:"timestamp"`
}

type IncomingMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type FCMMessage struct {
	Message FCMMessageContent `json:"message"`
}

type FCMMessageContent struct {
	Token        string        `json:"token,omitempty"`        // Target device token
	Notification *Notification `json:"notification,omitempty"` // Notification title and body
	Data         interface{}   `json:"data,omitempty"`         // Custom data payload
}

type Notification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}