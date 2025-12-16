package chat

// Chat defines the structure of a chat
type Chat struct {
	ID        string   `json:"id"`
	User1     string   `json:"user1"`
	User2     string   `json:"user2"`
	Messages  []string `json:"messages"`
	CreatedAt string   `json:"created_at"`
}
