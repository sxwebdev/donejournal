package processor

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/sxwebdev/donejournal/internal/bot"
	"github.com/sxwebdev/donejournal/internal/mcp"
	"github.com/sxwebdev/donejournal/internal/services/baseservices"
	"github.com/tkcrm/mx/logger"
)

type Processor struct {
	logger      logger.Logger
	baseService *baseservices.BaseServices
	mcpService  *mcp.MCP
	botService  *bot.Bot
}

func New(
	l logger.Logger,
	baseService *baseservices.BaseServices,
	mcpService *mcp.MCP,
	botService *bot.Bot,
) *Processor {
	s := &Processor{
		logger:      l,
		baseService: baseService,
		mcpService:  mcpService,
		botService:  botService,
	}

	return s
}

// inboxPrefixes are lowercased phrases that indicate the user wants to save directly to inbox.
var inboxPrefixes = []string{"в инбокс", "inbox", "инбокс", "in", "box"}

// parseInboxMessage returns (true, content) if the message starts with an inbox keyword.
// content is the text after the keyword, trimmed of leading spaces, colons, and newlines.
func parseInboxMessage(text string) (bool, string) {
	lower := strings.ToLower(strings.TrimSpace(text))
	for _, prefix := range inboxPrefixes {
		if strings.HasPrefix(lower, prefix) {
			rest := strings.TrimLeft(text[len(prefix):], " :\n\r\t")
			return true, rest
		}
	}
	return false, ""
}

func (s *Processor) ProcessNewRequest(ctx context.Context, userID int64, data string) (string, error) {
	if isInbox, content := parseInboxMessage(data); isInbox {
		if content == "" {
			content = data
		}
		if _, err := s.baseService.Inbox().Create(ctx, content, strconv.FormatInt(userID, 10)); err != nil {
			return "", fmt.Errorf("failed to create inbox item: %w", err)
		}
		return "Saved to inbox 📥", nil
	}

	resp, err := s.mcpService.ParseMessage(ctx, data)
	if err != nil {
		s.logger.Errorf("failed to parse message via MCP, saving to inbox: %v", err)
		if _, inboxErr := s.baseService.Inbox().Create(ctx, data, strconv.FormatInt(userID, 10)); inboxErr != nil {
			return "", fmt.Errorf("mcp error: %w; inbox fallback also failed: %v", err, inboxErr)
		}
		return "Saved to inbox 📥", nil
	}

	// Resolve project IDs from project names
	projectIDs := make(map[string]*string) // projectName -> projectID
	for _, entry := range resp.Entries {
		if entry.Project == "" {
			continue
		}
		if _, ok := projectIDs[entry.Project]; ok {
			continue
		}
		project, err := s.baseService.Projects().FindOrCreateByName(ctx, userID, entry.Project)
		if err != nil {
			s.logger.Warnw("failed to resolve project, ignoring", "project", entry.Project, "error", err)
			continue
		}
		projectIDs[entry.Project] = &project.ID
	}

	// Separate note entries from todo/done entries
	var todoEntries []mcp.ParsedEntry
	var noteEntries []mcp.ParsedEntry
	for _, entry := range resp.Entries {
		if entry.Kind == mcp.EntryKindNote {
			noteEntries = append(noteEntries, entry)
		} else {
			todoEntries = append(todoEntries, entry)
		}
	}

	// Create todos
	if len(todoEntries) > 0 {
		todoResp := &mcp.ParsedResponse{Entries: todoEntries}
		if err := s.baseService.Todos().BatchCreate(ctx, userID, todoResp, projectIDs); err != nil {
			return "", fmt.Errorf("failed to batch create todos: %w", err)
		}
	}

	// Create notes
	for _, entry := range noteEntries {
		body := entry.Body
		if body == "" {
			body = entry.Description
		}
		var projectID *string
		if entry.Project != "" {
			projectID = projectIDs[entry.Project]
		}
		if _, err := s.baseService.Notes().Create(ctx, userID, entry.Title, body, projectID); err != nil {
			return "", fmt.Errorf("failed to create note '%s': %w", entry.Title, err)
		}
	}

	if len(resp.Entries) == 0 {
		if _, err := s.baseService.Inbox().Create(ctx, data, strconv.FormatInt(userID, 10)); err != nil {
			return "", fmt.Errorf("failed to create inbox item: %w", err)
		}
		return "Saved to inbox 📥", nil
	}

	if len(resp.Entries) == 1 {
		entry := resp.Entries[0]
		prefix := "✅"
		switch entry.Kind {
		case mcp.EntryKindTodo:
			prefix = "📝"
		case mcp.EntryKindNote:
			prefix = "🗒️"
		}
		return fmt.Sprintf("%s %s", prefix, entry.Title), nil
	}

	responseText := new(strings.Builder)
	if _, err := fmt.Fprintf(responseText, "%d items have been created.\n", len(resp.Entries)); err != nil {
		return "", fmt.Errorf("failed to write result text: %w", err)
	}
	for i, entry := range resp.Entries {
		formatPreffix := "\n✅"
		switch entry.Kind {
		case mcp.EntryKindTodo:
			formatPreffix = "\n📝"
		case mcp.EntryKindNote:
			formatPreffix = "\n🗒️"
		}
		if _, err := fmt.Fprintf(responseText, formatPreffix+" %d. %s", i+1, entry.Title); err != nil {
			return "", fmt.Errorf("failed to write result text: %w", err)
		}
	}
	return responseText.String(), nil
}
