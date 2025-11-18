package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sxwebdev/donejournal/internal/mcp/mcptypes"
	"github.com/tkcrm/mx/logger"
)

type EntryKind string

const (
	EntryKindTodo EntryKind = "todo"
	EntryKindDone EntryKind = "done"
)

// ParsedEntry represents a single parsed task entry from the LLM
type ParsedEntry struct {
	Kind        EntryKind `json:"kind"`        // "todo" or "done"
	Title       string    `json:"title"`       // short title
	Date        string    `json:"date"`        // YYYY-MM-DD
	Description string    `json:"description"` // cleaned text
	Language    string    `json:"language"`    // "ru" | "en" | "unknown"
	Confidence  float64   `json:"confidence"`  // 0..1
}

// ParsedResponse contains all parsed entries from the LLM
type ParsedResponse struct {
	Entries []ParsedEntry `json:"entries"`
}

// MCP represents the MCP service with a specific provider
type MCP struct {
	provider mcptypes.MCPProvider
	log      logger.Logger
}

// New creates a new MCP instance with the given provider and logger
// Example usage:
//
//	provider := groq.NewClient(log, apiKey, model)
//	mcpService := mcp.New(log, provider)
//	response, err := mcpService.ParseMessage(ctx, userID, text)
func New(log logger.Logger, provider mcptypes.MCPProvider) *MCP {
	return &MCP{
		provider: provider,
		log:      log,
	}
}

// addDaysToDate adds a number of days to a date string in YYYY-MM-DD format
func addDaysToDate(dateStr string, days int) string {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}
	return t.AddDate(0, 0, days).Format("2006-01-02")
}

// ParseMessage processes user message through LLM provider and returns parsed entries
func (m *MCP) ParseMessage(ctx context.Context, text string) (*ParsedResponse, error) {
	m.log.Debugw(
		"mcp parse message",
		"text", text,
	)

	now := time.Now()
	today := now.Format("2006-01-02")

	systemPrompt := fmt.Sprintf(`Parse tasks from user messages. Today: %s

CRITICAL RULES:
1. If user lists MULTIPLE tasks, create SEPARATE entry for EACH task
2. If date word appears at the START of message (like "вчера", "yesterday", "завтра"), apply that date to ALL tasks in the message
3. If no date mentioned, use today's date for all tasks

Date words:
- "завтра"/"tomorrow" = %s
- "послезавтра"/"day after tomorrow" = %s  
- "вчера"/"yesterday" = %s
- "позавчера"/"day before yesterday" = %s
- no date = %s

Task type:
- "сделал"/"done"/"completed" -> kind:"done"
- "нужно"/"надо"/"добавь" -> kind:"todo"

IMPORTANT: Keep original language in title and description

Format: {"entries":[...]} where each entry is a separate task`,
		today,
		addDaysToDate(today, 1),
		addDaysToDate(today, 2),
		addDaysToDate(today, -1),
		addDaysToDate(today, -2),
		today)

	yesterday := addDaysToDate(today, -1)

	// Few-shot examples - showing EXACTLY how to split multiple tasks
	messages := []mcptypes.ChatMessage{
		{Role: "system", Content: systemPrompt},
		// Example 1: Multiple tasks with date at the start - ALL get the same date
		{Role: "user", Content: "вчера\n - таску A\n - таску B\n - таску C"},
		{Role: "assistant", Content: fmt.Sprintf(`{"entries":[{"kind":"done","title":"таску A","date":"%s","description":"таску A","language":"ru","confidence":0.95},{"kind":"done","title":"таску B","date":"%s","description":"таску B","language":"ru","confidence":0.95},{"kind":"done","title":"таску C","date":"%s","description":"таску C","language":"ru","confidence":0.95}]}`, yesterday, yesterday, yesterday)},
		// Example 2: Multiple tasks without date - all get today
		{Role: "user", Content: "сделал X, Y и Z"},
		{Role: "assistant", Content: fmt.Sprintf(`{"entries":[{"kind":"done","title":"X","date":"%s","description":"X","language":"ru","confidence":0.95},{"kind":"done","title":"Y","date":"%s","description":"Y","language":"ru","confidence":0.95},{"kind":"done","title":"Z","date":"%s","description":"Z","language":"ru","confidence":0.95}]}`, today, today, today)},
		// Example 3: Single task with date
		{Role: "user", Content: "добавь задачу на завтра - обновить домен"},
		{Role: "assistant", Content: fmt.Sprintf(`{"entries":[{"kind":"todo","title":"обновить домен","date":"%s","description":"обновить домен","language":"ru","confidence":0.97}]}`, addDaysToDate(today, 1))},
		// Actual user input
		{Role: "user", Content: text},
	}

	// Send request to LLM provider
	responseContent, err := m.provider.SendChatCompletion(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("provider request failed: %w", err)
	}

	// Parse response as JSON
	var parsed ParsedResponse
	if err := json.Unmarshal([]byte(responseContent), &parsed); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if len(parsed.Entries) > 0 {
		m.log.Infof("Successfully parsed %d entries from LLM response", len(parsed.Entries))
	}

	return &parsed, nil
}
