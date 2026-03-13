package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dromara/carbon/v2"
	"github.com/sxwebdev/donejournal/internal/agent/provider"
	"github.com/sxwebdev/donejournal/internal/services/baseservices"
	"github.com/sxwebdev/donejournal/internal/store/badgerdb"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_workspaces"
	"github.com/tkcrm/mx/logger"
)

const maxToolIterations = 10

// Agent orchestrates LLM tool-use conversations.
type Agent struct {
	provider     provider.Provider
	executor     *Executor
	conversation *ConversationStore
	services     *baseservices.BaseServices
	tools        []provider.ToolDefinition
	log          logger.Logger
}

// New creates a new Agent.
func New(
	log logger.Logger,
	llmProvider provider.Provider,
	services *baseservices.BaseServices,
	badgerDB *badgerdb.DB,
) *Agent {
	return &Agent{
		provider:     llmProvider,
		executor:     NewExecutor(services),
		conversation: NewConversationStore(badgerDB),
		services:     services,
		tools:        toolDefinitions(),
		log:          log,
	}
}

// Process handles a user message: loads conversation history, runs the agent loop,
// and returns the final text response.
func (a *Agent) Process(ctx context.Context, userID int64, text string) (string, error) {
	now := time.Now()

	// Load conversation history
	history, err := a.conversation.Load(ctx, userID)
	if err != nil {
		a.log.Warnw("failed to load conversation history, starting fresh", "error", err)
		history = nil
	}

	// Load user's workspaces for context
	var workspaceNames []string
	pageSize := uint32(50)
	if wsResult, err := a.services.Workspaces().Find(ctx, repo_workspaces.FindParams{
		UserID:   userID,
		PageSize: &pageSize,
	}); err == nil {
		for _, ws := range wsResult.Items {
			if !ws.Archived {
				workspaceNames = append(workspaceNames, ws.Name)
			}
		}
	}

	// Build messages: system prompt + history + new user message
	messages := make([]provider.ChatMessage, 0, len(history)+2)
	messages = append(messages, provider.ChatMessage{
		Role:    "system",
		Content: buildSystemPrompt(workspaceNames),
	})
	messages = append(messages, history...)
	messages = append(messages, provider.ChatMessage{
		Role:    "user",
		Content: text,
	})

	// Agent loop: call LLM, execute tools, repeat until text response
	var finalContent string
	for i := range maxToolIterations {
		resp, err := a.provider.ChatCompletion(ctx, provider.ChatRequest{
			Messages: messages,
			Tools:    a.tools,
		})
		if err != nil {
			return "", fmt.Errorf("llm request failed: %w", err)
		}

		// If no tool calls, we have the final response
		if len(resp.ToolCalls) == 0 {
			finalContent = resp.Content
			// Append assistant response to messages for history
			messages = append(messages, provider.ChatMessage{
				Role:    "assistant",
				Content: resp.Content,
			})
			break
		}

		a.log.Debugw("agent tool calls",
			"iteration", i+1,
			"tool_count", len(resp.ToolCalls),
		)

		// Append assistant message with tool calls
		messages = append(messages, provider.ChatMessage{
			Role:      "assistant",
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		})

		// Execute each tool call and append results
		for _, tc := range resp.ToolCalls {
			result, err := a.executor.Execute(ctx, userID, tc)
			if err != nil {
				a.log.Warnw("tool execution failed",
					"tool", tc.Function.Name,
					"error", err,
				)
				result = fmt.Sprintf(`{"error": "%s"}`, err.Error())
			}

			messages = append(messages, provider.ChatMessage{
				Role:       "tool",
				Content:    result,
				ToolCallID: tc.ID,
			})
		}
	}

	if finalContent == "" {
		finalContent = "Не удалось обработать запрос. Попробуйте ещё раз."
	}

	// Save conversation history (exclude system prompt)
	historyToSave := messages[1:] // skip system prompt
	if err := a.conversation.Save(ctx, userID, historyToSave); err != nil {
		a.log.Warnw("failed to save conversation history", "error", err)
	}

	a.log.Infow("agent request processed",
		"user_id", userID,
		"duration", time.Since(now).String(),
	)

	return finalContent, nil
}

// ClearConversation clears the conversation history for a user.
func (a *Agent) ClearConversation(ctx context.Context, userID int64) error {
	return a.conversation.Clear(ctx, userID)
}

func buildSystemPrompt(workspaceNames []string) string {
	today := carbon.Now()
	weekStart := carbon.Now().StartOfWeek()

	workspacesInfo := "User has no workspaces yet."
	if len(workspaceNames) > 0 {
		workspacesInfo = fmt.Sprintf("User's existing workspaces: %s. Use exact names when matching.", strings.Join(workspaceNames, ", "))
	}

	return fmt.Sprintf(`You are DoneJournal assistant. You help users manage their tasks (todos) and notes.

Today: %s (%s)
This week: Mon %s – Sun %s
%s

Rules:
- Respond in the user's language (detect from their message)
- To create tasks use create_todo tool
- To create notes use create_note tool
- To find/list tasks use find_todos tool
- To find/list notes use find_notes tool
- If user lists multiple tasks in one message, create each one separately with individual create_todo calls
- "сделал"/"done"/"completed" → create_todo with status="completed"
- "нужно"/"надо"/"добавь"/"todo" → create_todo with status="pending"
- "заметка"/"note"/"запомни" → create_note
- Handle relative dates: "вчера"/"yesterday", "завтра"/"tomorrow", "послезавтра", etc.
- Handle weekday references: "в понедельник"/"on Monday", "в прошлую среду"/"last Wednesday", etc.
- If workspace/project is mentioned ("для проекта X", "project X"), pass it to the workspace parameter
- When user mentions a workspace, match it to an existing one if possible (case-insensitive). Only create new workspaces when clearly intended.
- Keep the original language in task titles and descriptions
- Be concise in responses, use emoji for clarity
- When showing task lists, format them as numbered lists
- If the user's intent is unclear, ask a clarifying question instead of guessing
- When user asks to modify a task but doesn't specify which one, first use find_todos to find candidates, then ask user to clarify
- NEVER delete tasks or notes without explicit user confirmation. When user asks to delete, first show what will be deleted and ask "Are you sure?"
- When creating multiple items, respond with a compact numbered list of what was created`,
		today.ToDateString(),
		today.ToWeekString(),
		weekStart.ToDateString(),
		weekStart.AddDays(6).ToDateString(),
		workspacesInfo,
	)
}
