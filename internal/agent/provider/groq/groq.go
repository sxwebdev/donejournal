package groq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sxwebdev/donejournal/internal/agent/provider"
	"github.com/tkcrm/mx/logger"
)

// Client implements the provider.Provider interface for Groq API with tool-use support.
type Client struct {
	apiKey string
	model  string
	http   *http.Client
	log    logger.Logger
}

// NewClient creates a new Groq API client.
func NewClient(log logger.Logger, apiKey, model string) *Client {
	return &Client{
		apiKey: apiKey,
		model:  model,
		http:   &http.Client{Timeout: 120 * time.Second},
		log:    log,
	}
}

// ChatCompletion sends a chat completion request to Groq API with optional tool definitions.
func (c *Client) ChatCompletion(ctx context.Context, req provider.ChatRequest) (*provider.ChatResponse, error) {
	groqMessages := make([]groqChatMessage, len(req.Messages))
	for i, msg := range req.Messages {
		groqMessages[i] = groqChatMessage{
			Role:       msg.Role,
			Content:    msg.Content,
			ToolCallID: msg.ToolCallID,
		}
		if len(msg.ToolCalls) > 0 {
			groqMessages[i].ToolCalls = make([]groqToolCall, len(msg.ToolCalls))
			for j, tc := range msg.ToolCalls {
				groqMessages[i].ToolCalls[j] = groqToolCall{
					ID:   tc.ID,
					Type: tc.Type,
					Function: groqFunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}
		}
	}

	reqBody := groqRequest{
		Model:             c.model,
		Messages:          groqMessages,
		Temperature:       0.1,
		MaxTokens:         4096,
		TopP:              1,
		Stream:            false,
		ParallelToolCalls: true,
	}

	if len(req.Tools) > 0 {
		groqTools := make([]groqToolDefinition, len(req.Tools))
		for i, t := range req.Tools {
			groqTools[i] = groqToolDefinition{
				Type: t.Type,
				Function: groqFunctionDef{
					Name:        t.Function.Name,
					Description: t.Function.Description,
					Parameters:  t.Function.Parameters,
				},
			}
		}
		reqBody.Tools = groqTools
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal groq request: %w", err)
	}

	url := "https://api.groq.com/openai/v1/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("groq request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read groq response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("groq API error: status %d, body: %s", resp.StatusCode, string(respBytes))
	}

	var groqResp groqResponse
	if err := json.Unmarshal(respBytes, &groqResp); err != nil {
		return nil, fmt.Errorf("decode groq response: %w", err)
	}

	if len(groqResp.Choices) == 0 {
		return nil, fmt.Errorf("groq returned no choices")
	}

	choice := groqResp.Choices[0]
	result := &provider.ChatResponse{
		Content:      choice.Message.Content,
		FinishReason: choice.FinishReason,
	}

	if len(choice.Message.ToolCalls) > 0 {
		result.ToolCalls = make([]provider.ToolCall, len(choice.Message.ToolCalls))
		for i, tc := range choice.Message.ToolCalls {
			result.ToolCalls[i] = provider.ToolCall{
				ID:   tc.ID,
				Type: tc.Type,
				Function: provider.FunctionCall{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
		}
	}

	return result, nil
}

// Groq API types

type groqRequest struct {
	Model             string               `json:"model"`
	Messages          []groqChatMessage    `json:"messages"`
	Temperature       float64              `json:"temperature"`
	MaxTokens         int                  `json:"max_completion_tokens"`
	TopP              float64              `json:"top_p"`
	Stream            bool                 `json:"stream"`
	ParallelToolCalls bool                 `json:"parallel_tool_calls"`
	Tools             []groqToolDefinition `json:"tools,omitempty"`
}

type groqChatMessage struct {
	Role       string         `json:"role"`
	Content    string         `json:"content"`
	ToolCalls  []groqToolCall `json:"tool_calls,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
}

type groqToolDefinition struct {
	Type     string          `json:"type"`
	Function groqFunctionDef `json:"function"`
}

type groqFunctionDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type groqToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function groqFunctionCall `json:"function"`
}

type groqFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
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
