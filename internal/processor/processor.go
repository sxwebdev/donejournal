package processor

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/sxwebdev/donejournal/internal/agent"
	"github.com/sxwebdev/donejournal/internal/bot"
	"github.com/sxwebdev/donejournal/internal/services/baseservices"
	"github.com/tkcrm/mx/logger"
)

type Processor struct {
	logger      logger.Logger
	baseService *baseservices.BaseServices
	agent       *agent.Agent
	botService  *bot.Bot
}

func New(
	l logger.Logger,
	baseService *baseservices.BaseServices,
	agentService *agent.Agent,
	botService *bot.Bot,
) *Processor {
	return &Processor{
		logger:      l,
		baseService: baseService,
		agent:       agentService,
		botService:  botService,
	}
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
	// Direct inbox save
	if isInbox, content := parseInboxMessage(data); isInbox {
		if content == "" {
			content = data
		}
		if _, err := s.baseService.Inbox().Create(ctx, content, strconv.FormatInt(userID, 10)); err != nil {
			return "", fmt.Errorf("failed to create inbox item: %w", err)
		}
		return "Saved to inbox 📥", nil
	}

	// Delegate to agent
	response, err := s.agent.Process(ctx, userID, data)
	if err != nil {
		s.logger.Errorf("agent failed, saving to inbox: %v", err)
		if _, inboxErr := s.baseService.Inbox().Create(ctx, data, strconv.FormatInt(userID, 10)); inboxErr != nil {
			return "", fmt.Errorf("agent error: %w; inbox fallback also failed: %v", err, inboxErr)
		}
		return "Saved to inbox 📥", nil
	}

	return response, nil
}
