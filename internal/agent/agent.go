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
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_tags"
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

	// Load user's tags for context
	var tagNames []string
	if tagResult, err := a.services.Tags().Find(ctx, repo_tags.FindParams{
		UserID:   userID,
		PageSize: &pageSize,
	}); err == nil {
		for _, t := range tagResult.Items {
			tagNames = append(tagNames, t.Name)
		}
	}

	// Build messages: system prompt + history + ephemeral date reminder + new user message.
	// The ephemeral date reminder is placed adjacent to the user message to keep the current
	// date salient — without it, dates in stale tool calls from history bias the model.
	messages := make([]provider.ChatMessage, 0, len(history)+3)
	messages = append(messages, provider.ChatMessage{
		Role:    "system",
		Content: buildSystemPrompt(workspaceNames, tagNames),
	})
	messages = append(messages, history...)
	messages = append(messages, provider.ChatMessage{
		Role:    "system",
		Content: fmt.Sprintf("Current date and time: %s %s (%s). Use this for any 'today' / 'now' / relative date references in this turn. Ignore any dates from previous tool calls in history — those were created on different days.", carbon.Now().ToDateString(), carbon.Now().ToTimeString(), carbon.Now().Timezone()),
	})
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
			a.log.Infow("agent tool call",
				"iteration", i+1,
				"tool", tc.Function.Name,
				"args", tc.Function.Arguments,
			)
			result, err := a.executor.Execute(ctx, userID, tc)
			if err != nil {
				a.log.Warnw("tool execution failed",
					"tool", tc.Function.Name,
					"error", err,
				)
				result = fmt.Sprintf(`{"error": "%s"}`, err.Error())
			} else {
				a.log.Debugw("agent tool result",
					"tool", tc.Function.Name,
					"result", result,
				)
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

	// Save conversation history: only user messages and final assistant text.
	// Tool calls / tool results are intra-turn mechanics — persisting them pollutes
	// future contexts (especially with stale dates in arguments) without adding value,
	// since tool_call_ids are one-shot and the user-visible answer captures the outcome.
	historyToSave := make([]provider.ChatMessage, 0, len(messages))
	for _, m := range messages {
		if m.Role == "system" || m.Role == "tool" {
			continue
		}
		if m.Role == "assistant" && len(m.ToolCalls) > 0 {
			continue
		}
		historyToSave = append(historyToSave, m)
	}
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

func buildSystemPrompt(workspaceNames []string, tagNames []string) string {
	today := carbon.Now()
	weekStart := carbon.Now().StartOfWeek()

	workspacesInfo := "User has no workspaces yet."
	if len(workspaceNames) > 0 {
		workspacesInfo = fmt.Sprintf("User's existing workspaces: %s. Use exact names when matching.", strings.Join(workspaceNames, ", "))
	}

	tagsInfo := "User has no tags yet."
	if len(tagNames) > 0 {
		tagsInfo = fmt.Sprintf("User's existing tags: %s. Use exact names when matching.", strings.Join(tagNames, ", "))
	}

	return fmt.Sprintf(`You are DoneJournal assistant. You help users manage their tasks (todos) and notes.

Today: %s (%s), current time: %s (%s)
This week: Mon %s – Sun %s
%s
%s

Rules:
- Respond in the user's language (detect from their message)
- CRITICAL — NEVER fabricate task data. If the user asks anything about tasks (count, list, search, delete, "сколько", "найди", "удали", "выполнены"), you MUST call find_todos FIRST and base your answer on its result. Do not reuse counts, IDs, or task lists from earlier turns in this conversation — they may be stale or for different filters. Saying "найдено 0 задач" without a fresh find_todos call is forbidden.
- To create tasks use create_todo tool
- To create notes use create_note tool
- To find/list tasks use find_todos tool
- To find/list notes use find_notes tool
- If user lists multiple tasks in one message, create each one separately with individual create_todo calls
- "сделал"/"done"/"completed" → create_todo with status="completed"
- "нужно"/"надо"/"добавь"/"todo" → create_todo with status="pending"
- "заметка"/"note"/"запомни" → create_note
- ALWAYS use the "Today" value above to resolve relative dates ("сегодня", "завтра", "вчера", "на этой неделе"). Never assume "today" is a date you saw earlier in the conversation.
- DO NOT copy planned_date values from previous tool calls in conversation history — those tasks were created on different days and have no bearing on the current date.
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
- Bulk delete: if user asks to delete MULTIPLE tasks by criteria (e.g. "удали все завершённые до DATE", "delete cancelled tasks in project X"): (1) call find_todos with the same filters and explicit status, (2) show the count and a short numbered list, ask for confirmation in the user's language, (3) on confirmation call bulk_delete_todos with the same filters and confirmed=true. Do NOT loop delete_todo for bulk operations — it will hit the iteration limit.
- Date bound semantics: by default "до DATE" / "по DATE" / "until DATE" / "through DATE" is INCLUSIVE — pass date_to=DATE. Only treat as exclusive when user is explicit ("before DATE", "до DATE не включая", "строго раньше DATE") — then pass date_to as the day before DATE. Echo the actual date you used in your reply so the user can verify ("до 2026-04-27 включительно").
- When filtering completed tasks by completion date, set status=["completed"] so date_from/date_to filter by completed_at instead of planned_date.
- When creating multiple items, respond with a compact numbered list of what was created
- Priority keywords: "critical"/"срочно"/"asap"/"критично" → priority=critical, "important"/"важно" → priority=high, "medium priority"/"средний приоритет" → priority=medium, "low priority"/"неважно"/"когда-нибудь" → priority=low
- Default priority is none unless user explicitly mentions importance or urgency
- If user mentions #hashtags (e.g. "#urgent купить молоко"), extract them as tag names and pass in the tags parameter
- When user mentions tags, match to existing ones if possible (case-insensitive). New tags are auto-created.
- Use manage_tags to list/create/delete tags, tag_entity to add tags to existing items, find_by_tag to search by tags
- Recurring tasks: use create_recurring_todo when user says "every day"/"каждый день", "every week"/"каждую неделю"/"еженедельно", "every month"/"каждый месяц"/"ежемесячно", "every Monday"/"каждый понедельник", etc.
- recurrence_rule values: "daily" for daily/every day, "weekly" for weekly/every week/every weekday, "monthly" for monthly
- When a recurring todo is completed, the next occurrence is created automatically`,
		today.ToDateString(),
		today.ToWeekString(),
		today.ToTimeString(),
		today.Timezone(),
		weekStart.ToDateString(),
		weekStart.AddDays(6).ToDateString(),
		workspacesInfo,
		tagsInfo,
	)
}
