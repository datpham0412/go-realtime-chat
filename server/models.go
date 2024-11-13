package server

// Message represents a chat message
type Message struct {
	ID        string `json:"id"`
	User      string `json:"user"`
	Text      string `json:"text"`
	CreatedAt Time   `json:"createdAt"`
} 