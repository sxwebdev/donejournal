package aimlapi

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

const apiURL = "https://api.aimlapi.com/v1/chat/completions"

// Provider implements the provider.Provider interface for the AIML API
// (https://docs.aimlapi.com) — an OpenAI-compatible gateway exposing many
// third-party models behind a single chat completion endpoint.
type Provider struct {
	apiKey string
	model  string
	http   *http.Client
	log    logger.Logger
}

// NewProvider creates a new AIML API client.
func NewProvider(log logger.Logger, apiKey, model string) *Provider {
	return &Provider{
		apiKey: apiKey,
		model:  model,
		http:   &http.Client{Timeout: 120 * time.Second},
		log:    log,
	}
}

// ChatCompletion sends a chat completion request to AIML API.
func (p *Provider) ChatCompletion(ctx context.Context, req provider.ChatRequest) (*provider.ChatResponse, error) {
	body, err := buildBody(p.model, req)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("aimlapi marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("aimlapi create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("aimlapi request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("aimlapi read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("aimlapi API error: status %d, body: %s", resp.StatusCode, string(respBytes))
	}

	var parsed openAIChatResponse
	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		return nil, fmt.Errorf("aimlapi decode response: %w", err)
	}

	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("aimlapi returned no choices")
	}

	choice := parsed.Choices[0]
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

func buildBody(model string, req provider.ChatRequest) (map[string]any, error) {
	messages := make([]map[string]any, len(req.Messages))
	for i, m := range req.Messages {
		msg := map[string]any{
			"role":    m.Role,
			"content": m.Content,
		}
		if m.ToolCallID != "" {
			msg["tool_call_id"] = m.ToolCallID
		}
		if len(m.ToolCalls) > 0 {
			calls := make([]map[string]any, len(m.ToolCalls))
			for j, tc := range m.ToolCalls {
				calls[j] = map[string]any{
					"id":   tc.ID,
					"type": tc.Type,
					"function": map[string]any{
						"name":      tc.Function.Name,
						"arguments": tc.Function.Arguments,
					},
				}
			}
			msg["tool_calls"] = calls
		}
		messages[i] = msg
	}

	body := map[string]any{
		"model":               model,
		"messages":            messages,
		"temperature":         0.1,
		"max_tokens":          4096,
		"top_p":               1,
		"stream":              false,
		"parallel_tool_calls": true,
	}

	if len(req.Tools) > 0 {
		tools := make([]map[string]any, len(req.Tools))
		for i, t := range req.Tools {
			var params map[string]any
			if len(t.Function.Parameters) > 0 {
				if err := json.Unmarshal(t.Function.Parameters, &params); err != nil {
					return nil, fmt.Errorf("aimlapi: tool %q parameters: %w", t.Function.Name, err)
				}
			}
			tools[i] = map[string]any{
				"type": t.Type,
				"function": map[string]any{
					"name":        t.Function.Name,
					"description": t.Function.Description,
					"parameters":  params,
				},
			}
		}
		body["tools"] = tools
	}

	return body, nil
}

// openAIChatResponse mirrors the OpenAI-compatible chat completion envelope
// returned by AIML API.
type openAIChatResponse struct {
	Choices []chatChoice `json:"choices"`
}

type chatChoice struct {
	Index        int         `json:"index"`
	Message      chatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type chatMessage struct {
	Role      string         `json:"role"`
	Content   string         `json:"content"`
	ToolCalls []chatToolCall `json:"tool_calls,omitempty"`
}

type chatToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function chatFunctionCall `json:"function"`
}

type chatFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}
