package openrouter

import (
	"context"
	"encoding/json"
	"fmt"

	openrouter "github.com/OpenRouterTeam/go-sdk"
	"github.com/OpenRouterTeam/go-sdk/models/components"
	"github.com/OpenRouterTeam/go-sdk/optionalnullable"

	"github.com/sxwebdev/donejournal/internal/agent/provider"
	"github.com/tkcrm/mx/logger"
)

// Client implements the provider.Provider interface for the OpenRouter API
// using the official OpenRouter Go SDK.
type Client struct {
	sdk   *openrouter.OpenRouter
	model string
	log   logger.Logger
}

// NewClient creates a new OpenRouter API client.
func NewClient(log logger.Logger, apiKey, model string) *Client {
	return &Client{
		sdk:   openrouter.New(openrouter.WithSecurity(apiKey)),
		model: model,
		log:   log,
	}
}

// ChatCompletion sends a chat completion request to OpenRouter with optional tool definitions.
func (c *Client) ChatCompletion(ctx context.Context, req provider.ChatRequest) (*provider.ChatResponse, error) {
	messages, err := buildMessages(req.Messages)
	if err != nil {
		return nil, err
	}

	chatReq := components.ChatRequest{
		Model:       openrouter.Pointer(c.model),
		Messages:    messages,
		Temperature: optionalnullable.From(openrouter.Pointer[float64](0.1)),
		MaxTokens:   optionalnullable.From(openrouter.Pointer[int64](4096)),
		TopP:        optionalnullable.From(openrouter.Pointer[float64](1)),
		Stream:      openrouter.Pointer(false),
	}

	if len(req.Tools) > 0 {
		tools, err := buildTools(req.Tools)
		if err != nil {
			return nil, err
		}
		chatReq.Tools = tools
	}

	res, err := c.sdk.Chat.Send(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("openrouter chat send: %w", err)
	}
	if res == nil || res.ChatResult == nil {
		return nil, fmt.Errorf("openrouter empty response")
	}

	choices := res.ChatResult.Choices
	if len(choices) == 0 {
		return nil, fmt.Errorf("openrouter returned no choices")
	}

	choice := choices[0]
	result := &provider.ChatResponse{}

	if choice.FinishReason != nil {
		result.FinishReason = string(*choice.FinishReason)
	}

	if content, ok := choice.Message.Content.Get(); ok && content != nil && content.Str != nil {
		result.Content = *content.Str
	}

	if len(choice.Message.ToolCalls) > 0 {
		result.ToolCalls = make([]provider.ToolCall, len(choice.Message.ToolCalls))
		for i, tc := range choice.Message.ToolCalls {
			result.ToolCalls[i] = provider.ToolCall{
				ID:   tc.ID,
				Type: string(tc.Type),
				Function: provider.FunctionCall{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
		}
	}

	return result, nil
}

func buildMessages(msgs []provider.ChatMessage) ([]components.ChatMessages, error) {
	out := make([]components.ChatMessages, 0, len(msgs))
	for i, m := range msgs {
		switch m.Role {
		case "system":
			out = append(out, components.CreateChatMessagesSystem(components.ChatSystemMessage{
				Content: components.CreateChatSystemMessageContentStr(m.Content),
			}))
		case "user":
			out = append(out, components.CreateChatMessagesUser(components.ChatUserMessage{
				Content: components.CreateChatUserMessageContentStr(m.Content),
			}))
		case "assistant":
			am := components.ChatAssistantMessage{}
			if m.Content != "" {
				am.Content = optionalnullable.From(openrouter.Pointer(
					components.CreateChatAssistantMessageContentStr(m.Content),
				))
			}
			if len(m.ToolCalls) > 0 {
				am.ToolCalls = make([]components.ChatToolCall, len(m.ToolCalls))
				for j, tc := range m.ToolCalls {
					am.ToolCalls[j] = components.ChatToolCall{
						ID:   tc.ID,
						Type: components.ChatToolCallTypeFunction,
						Function: components.ChatToolCallFunction{
							Name:      tc.Function.Name,
							Arguments: tc.Function.Arguments,
						},
					}
				}
			}
			out = append(out, components.CreateChatMessagesAssistant(am))
		case "tool":
			out = append(out, components.CreateChatMessagesTool(components.ChatToolMessage{
				Content:    components.CreateChatToolMessageContentStr(m.Content),
				ToolCallID: m.ToolCallID,
			}))
		default:
			return nil, fmt.Errorf("openrouter: unsupported message role %q at index %d", m.Role, i)
		}
	}
	return out, nil
}

func buildTools(tools []provider.ToolDefinition) ([]components.ChatFunctionTool, error) {
	out := make([]components.ChatFunctionTool, len(tools))
	for i, t := range tools {
		params, err := decodeParameters(t.Function.Parameters)
		if err != nil {
			return nil, fmt.Errorf("openrouter: tool %q parameters: %w", t.Function.Name, err)
		}
		out[i] = components.CreateChatFunctionToolChatFunctionToolFunction(components.ChatFunctionToolFunction{
			Type: components.ChatFunctionToolTypeFunction,
			Function: components.ChatFunctionToolFunctionFunction{
				Name:        t.Function.Name,
				Description: openrouter.Pointer(t.Function.Description),
				Parameters:  params,
			},
		})
	}
	return out, nil
}

func decodeParameters(raw json.RawMessage) (map[string]any, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	return m, nil
}
