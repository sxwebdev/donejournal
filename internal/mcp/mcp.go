package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dromara/carbon/v2"
	"github.com/sxwebdev/donejournal/internal/mcp/mcptypes"
	"github.com/tkcrm/mx/logger"
)

type EntryKind string

const (
	EntryKindTodo EntryKind = "todo"
	EntryKindDone EntryKind = "done"
	EntryKindNote EntryKind = "note"
)

// ParsedEntry represents a single parsed task entry from the LLM
type ParsedEntry struct {
	Kind        EntryKind `json:"kind"`        // "todo", "done", or "note"
	Title       string    `json:"title"`       // short title
	Date        string    `json:"date"`        // YYYY-MM-DD
	Description string    `json:"description"` // cleaned text
	Body        string    `json:"body"`        // full text (for notes)
	Project     string    `json:"project"`     // optional project name
	Language    string    `json:"language"`     // "ru" | "en" | "unknown"
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

// addDaysToDate adds a number of days to current time and returns in YYYY-MM-DD HH:mm:ss format
func addDaysToDate(days int) string {
	return carbon.Now().AddDays(days).ToDateTimeString()
}

// weekDate returns the YYYY-MM-DD date for a day relative to the current week.
// weekOffset: 0=this week, -1=last week, +1=next week.
// dayOffset: 0=Mon, 1=Tue, 2=Wed, 3=Thu, 4=Fri, 5=Sat, 6=Sun.
// Note: carbon.AddDays mutates in place, so each call starts from a fresh carbon.Now().
func weekDate(weekOffset, dayOffset int) string {
	return carbon.Now().StartOfWeek().AddDays(weekOffset*7 + dayOffset).ToDateString()
}

// ParseMessage processes user message through LLM provider and returns parsed entries
func (m *MCP) ParseMessage(ctx context.Context, text string) (*ParsedResponse, error) {
	m.log.Debugw(
		"mcp parse message",
		"text", text,
	)

	now := time.Now()
	today := carbon.Now().ToDateTimeString()
	todayWeekday := carbon.Now().ToWeekString()
	thisWeekDates := fmt.Sprintf("%s, %s, %s, %s, %s, %s, %s",
		weekDate(0, 0), weekDate(0, 1), weekDate(0, 2), weekDate(0, 3),
		weekDate(0, 4), weekDate(0, 5), weekDate(0, 6))

	systemPrompt := fmt.Sprintf(`Parse tasks from user messages. Today: %s (%s)

CRITICAL RULES:
1. If user lists MULTIPLE tasks, create SEPARATE entry for EACH task
2. If date word appears at the START of message (like "вчера", "yesterday", "завтра"), apply that date to ALL tasks in the message
3. If no date mentioned, use today's date for all tasks

Date words:
- "завтра"/"tomorrow"                = %s
- "послезавтра"/"day after tomorrow" = %s
- "вчера"/"yesterday"               = %s
- "позавчера"/"day before yesterday" = %s
- no date                            = %s

This week Mon–Sun: %s
- "в [день]" / "в эту/этот [день]" → use date from this week above
- "в прошлую/прошлый [день]" → subtract 7 days from this week's date
- "в следующую/следующий [день]" → add 7 days to this week's date
Russian: понедельник=Mon, вторник=Tue, среду=Wed, четверг=Thu, пятницу=Fri, субботу=Sat, воскресенье=Sun

Task type:
- "сделал"/"done"/"completed" -> kind:"done"
- "нужно"/"надо"/"добавь" -> kind:"todo"
- "запиши"/"заметка"/"note"/"remember"/"запомни" -> kind:"note"

For notes: "title" is a short title, "body" is the full text content (markdown). Date is not used for notes.

Project: if user mentions a project name (e.g., "для проекта X", "project X", "в проект X", "проект: X"), extract the project name into "project" field. If no project mentioned, leave "project" as empty string.

IMPORTANT: Keep original language in title, description, and body

Format: {"entries":[...]} where each entry is a separate task or note`,
		today, todayWeekday,
		addDaysToDate(1),
		addDaysToDate(2),
		addDaysToDate(-1),
		addDaysToDate(-2),
		today,
		thisWeekDates)

	// Few-shot examples - showing EXACTLY how to split multiple tasks
	messages := []mcptypes.ChatMessage{
		{Role: "system", Content: systemPrompt},
		// Example 1: Multiple tasks with date at the start - ALL get the same date
		{Role: "user", Content: "вчера\n - таску A\n - таску B\n - таску C"},
		{Role: "assistant", Content: fmt.Sprintf(`{"entries":[{"kind":"done","title":"таску A","date":"%s","description":"таску A","language":"ru","confidence":0.95},{"kind":"done","title":"таску B","date":"%s","description":"таску B","language":"ru","confidence":0.95},{"kind":"done","title":"таску C","date":"%s","description":"таску C","language":"ru","confidence":0.95}]}`, addDaysToDate(-1), addDaysToDate(-1), addDaysToDate(-1))},
		// Example 2: Multiple tasks without date - all get today
		{Role: "user", Content: "сделал X, Y и Z"},
		{Role: "assistant", Content: fmt.Sprintf(`{"entries":[{"kind":"done","title":"X","date":"%s","description":"X","language":"ru","confidence":0.95},{"kind":"done","title":"Y","date":"%s","description":"Y","language":"ru","confidence":0.95},{"kind":"done","title":"Z","date":"%s","description":"Z","language":"ru","confidence":0.95}]}`, today, today, today)},
		// Example 3: Day-of-week — "в субботу"=this week, "в прошлую субботу"=-7 days
		{Role: "user", Content: "в субботу сделал рефакторинг, в прошлую субботу завершил дизайн"},
		{Role: "assistant", Content: fmt.Sprintf(`{"entries":[{"kind":"done","title":"рефакторинг","date":"%s","description":"рефакторинг","language":"ru","confidence":0.95},{"kind":"done","title":"дизайн","date":"%s","description":"дизайн","language":"ru","confidence":0.95}]}`, weekDate(0, 5), weekDate(-1, 5))},
		// Example 4: Single task with date
		{Role: "user", Content: "добавь задачу на завтра - обновить домен"},
		{Role: "assistant", Content: fmt.Sprintf(`{"entries":[{"kind":"todo","title":"обновить домен","date":"%s","description":"обновить домен","language":"ru","confidence":0.97}]}`, addDaysToDate(1))},
		// Example 5: Note creation
		{Role: "user", Content: "заметка: идеи для нового проекта\nИспользовать Go + React\nДобавить авторизацию через Telegram"},
		{Role: "assistant", Content: fmt.Sprintf(`{"entries":[{"kind":"note","title":"идеи для нового проекта","date":"%s","description":"","body":"Использовать Go + React\nДобавить авторизацию через Telegram","language":"ru","confidence":0.95}]}`, today)},
		// Example 6: Task with project
		{Role: "user", Content: "для проекта DoneJournal: добавь задачу на завтра - сделать фильтр по проектам"},
		{Role: "assistant", Content: fmt.Sprintf(`{"entries":[{"kind":"todo","title":"сделать фильтр по проектам","date":"%s","description":"сделать фильтр по проектам","project":"DoneJournal","language":"ru","confidence":0.95}]}`, addDaysToDate(1))},
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

	for i := range parsed.Entries {
		if parsed.Entries[i].Title == parsed.Entries[i].Description {
			parsed.Entries[i].Description = ""
		}
	}

	if len(parsed.Entries) > 0 {
		m.log.Infow(
			"Successfully parsed entries from LLM response",
			"count", len(parsed.Entries),
			"time", time.Since(now).String())
	}

	return &parsed, nil
}
