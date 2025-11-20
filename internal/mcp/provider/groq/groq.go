package groq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sxwebdev/donejournal/internal/mcp/mcptypes"
	"github.com/tkcrm/mx/logger"
)

// Client implements MCP Provider interface for Groq API
type Client struct {
	apiKey string
	model  string
	http   *http.Client
	log    logger.Logger
}

// NewClient creates a new Groq API client
func NewClient(log logger.Logger, apiKey, model string) *Client {
	return &Client{
		apiKey: apiKey,
		model:  model,
		http:   &http.Client{Timeout: 120 * time.Second},
		log:    log,
	}
}

// SendChatCompletion sends a chat completion request to Groq API
func (c *Client) SendChatCompletion(ctx context.Context, messages []mcptypes.ChatMessage) (string, error) {
	groqMessages := make([]groqChatMessage, len(messages))
	for i, msg := range messages {
		groqMessages[i] = groqChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	reqBody := groqRequest{
		Model:       c.model,
		Messages:    groqMessages,
		Temperature: 0.1,
		MaxTokens:   2048,
		TopP:        1,
		Stream:      false,
		Stop:        nil,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal groq request: %w", err)
	}

	url := "https://api.groq.com/openai/v1/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("groq request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read groq response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("groq API error: status %d, body: %s", resp.StatusCode, string(respBytes))
	}

	var groqResp groqResponse
	if err := json.Unmarshal(respBytes, &groqResp); err != nil {
		return "", fmt.Errorf("decode groq response: %w", err)
	}

	if len(groqResp.Choices) == 0 {
		return "", fmt.Errorf("groq returned no choices")
	}

	return groqResp.Choices[0].Message.Content, nil
}

// Internal Groq API structures
type groqRequest struct {
	Model          string            `json:"model"`
	Messages       []groqChatMessage `json:"messages"`
	Temperature    float64           `json:"temperature"`
	MaxTokens      int               `json:"max_completion_tokens"`
	TopP           float64           `json:"top_p"`
	Stream         bool              `json:"stream"`
	Stop           interface{}       `json:"stop"`
	ResponseFormat *responseFormat   `json:"response_format,omitempty"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type groqChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []choice `json:"choices"`
	Usage   usage    `json:"usage"`
}

type choice struct {
	Index        int             `json:"index"`
	Message      groqChatMessage `json:"message"`
	FinishReason string          `json:"finish_reason"`
}

type usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
