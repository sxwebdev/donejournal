package mcptypes

import "context"

// ChatMessage represents a single message in the chat conversation
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// MCPProvider is the interface that all LLM providers must implement
type MCPProvider interface {
	SendChatCompletion(ctx context.Context, messages []ChatMessage) (string, error)
}
