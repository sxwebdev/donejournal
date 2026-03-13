package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/dromara/carbon/v2"
	"github.com/sxwebdev/donejournal/internal/agent/provider"
	"github.com/sxwebdev/donejournal/internal/models"
	"github.com/sxwebdev/donejournal/internal/services/baseservices"
	"github.com/sxwebdev/donejournal/internal/services/notes"
	"github.com/sxwebdev/donejournal/internal/services/todos"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_notes"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_todos"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_workspaces"
)

// Executor dispatches tool calls to the appropriate service methods.
type Executor struct {
	services *baseservices.BaseServices
}

// NewExecutor creates a new Executor.
func NewExecutor(services *baseservices.BaseServices) *Executor {
	return &Executor{services: services}
}

// Execute runs a tool call and returns the JSON result string.
func (e *Executor) Execute(ctx context.Context, userID int64, call provider.ToolCall) (string, error) {
	switch call.Function.Name {
	case "create_todo":
		return e.createTodo(ctx, userID, call.Function.Arguments)
	case "create_note":
		return e.createNote(ctx, userID, call.Function.Arguments)
	case "find_todos":
		return e.findTodos(ctx, userID, call.Function.Arguments)
	case "find_notes":
		return e.findNotes(ctx, userID, call.Function.Arguments)
	case "complete_todo":
		return e.completeTodo(ctx, userID, call.Function.Arguments)
	case "update_todo":
		return e.updateTodo(ctx, userID, call.Function.Arguments)
	case "list_workspaces":
		return e.listWorkspaces(ctx, userID)
	case "save_to_inbox":
		return e.saveToInbox(ctx, userID, call.Function.Arguments)
	case "delete_todo":
		return e.deleteTodo(ctx, userID, call.Function.Arguments)
	case "delete_note":
		return e.deleteNote(ctx, userID, call.Function.Arguments)
	case "update_note":
		return e.updateNote(ctx, userID, call.Function.Arguments)
	case "get_todo_stats":
		return e.getTodoStats(ctx, userID, call.Function.Arguments)
	case "list_inbox":
		return e.listInbox(ctx, userID)
	case "convert_inbox":
		return e.convertInbox(ctx, userID, call.Function.Arguments)
	default:
		return "", fmt.Errorf("unknown tool: %s", call.Function.Name)
	}
}

// --- Tool implementations ---

type createTodoArgs struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	PlannedDate string `json:"planned_date"`
	Status      string `json:"status"`
	Workspace   string `json:"workspace"`
}

func (e *Executor) createTodo(ctx context.Context, userID int64, argsJSON string) (string, error) {
	var args createTodoArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse create_todo args: %w", err)
	}

	plannedDate := time.Now()
	if args.PlannedDate != "" {
		parsed := carbon.Parse(args.PlannedDate).StdTime()
		if !parsed.IsZero() {
			plannedDate = parsed
		}
	}

	var workspaceID *string
	if args.Workspace != "" {
		ws, err := e.services.Workspaces().FindOrCreateByName(ctx, userID, args.Workspace)
		if err != nil {
			return "", fmt.Errorf("resolve workspace: %w", err)
		}
		workspaceID = &ws.ID
	}

	todo, err := e.services.Todos().CreateFromAPI(ctx, userID, args.Title, args.Description, plannedDate, workspaceID)
	if err != nil {
		return "", fmt.Errorf("create todo: %w", err)
	}

	// If status is completed, mark it as completed
	if args.Status == "completed" {
		todo, err = e.services.Todos().Complete(ctx, userID, todo.ID)
		if err != nil {
			return "", fmt.Errorf("complete todo: %w", err)
		}
	}

	return toJSON(map[string]any{
		"id":           todo.ID,
		"title":        todo.Title,
		"status":       todo.Status,
		"planned_date": todo.PlannedDate.Format("2006-01-02"),
	})
}

type createNoteArgs struct {
	Title     string `json:"title"`
	Body      string `json:"body"`
	Workspace string `json:"workspace"`
}

func (e *Executor) createNote(ctx context.Context, userID int64, argsJSON string) (string, error) {
	var args createNoteArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse create_note args: %w", err)
	}

	var workspaceID *string
	if args.Workspace != "" {
		ws, err := e.services.Workspaces().FindOrCreateByName(ctx, userID, args.Workspace)
		if err != nil {
			return "", fmt.Errorf("resolve workspace: %w", err)
		}
		workspaceID = &ws.ID
	}

	note, err := e.services.Notes().Create(ctx, userID, args.Title, args.Body, workspaceID)
	if err != nil {
		return "", fmt.Errorf("create note: %w", err)
	}

	return toJSON(map[string]any{
		"id":    note.ID,
		"title": note.Title,
	})
}

type findTodosArgs struct {
	Status   []string `json:"status"`
	DateFrom string   `json:"date_from"`
	DateTo   string   `json:"date_to"`
	Workspace string  `json:"workspace"`
}

func (e *Executor) findTodos(ctx context.Context, userID int64, argsJSON string) (string, error) {
	var args findTodosArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse find_todos args: %w", err)
	}

	params := repo_todos.FindParams{
		UserID: userID,
	}

	if len(args.Status) > 0 {
		for _, s := range args.Status {
			params.Statuses = append(params.Statuses, models.TodoStatusType(s))
		}
	} else {
		params.Statuses = []models.TodoStatusType{models.TodoStatusPending, models.TodoStatusInProgress}
	}

	if args.DateFrom != "" {
		t := carbon.Parse(args.DateFrom).StartOfDay().StdTime()
		if !t.IsZero() {
			params.DateFrom = &t
		}
	}
	if args.DateTo != "" {
		t := carbon.Parse(args.DateTo).EndOfDay().StdTime()
		if !t.IsZero() {
			params.DateTo = &t
		}
	}

	if args.Workspace != "" {
		ws, err := e.services.Workspaces().FindOrCreateByName(ctx, userID, args.Workspace)
		if err == nil {
			params.WorkspaceID = &ws.ID
		}
	}

	pageSize := uint32(50)
	params.PageSize = &pageSize

	result, err := e.services.Todos().Find(ctx, params)
	if err != nil {
		return "", fmt.Errorf("find todos: %w", err)
	}

	items := make([]map[string]any, 0, len(result.Items))
	for _, todo := range result.Items {
		item := map[string]any{
			"id":           todo.ID,
			"title":        todo.Title,
			"status":       string(todo.Status),
			"planned_date": todo.PlannedDate.Format("2006-01-02"),
		}
		if todo.Description != "" {
			item["description"] = todo.Description
		}
		if todo.CompletedAt != nil {
			item["completed_at"] = todo.CompletedAt.Format("2006-01-02 15:04")
		}
		items = append(items, item)
	}

	return toJSON(map[string]any{
		"total_count": result.Count,
		"items":       items,
	})
}

type findNotesArgs struct {
	Search    string `json:"search"`
	Workspace string `json:"workspace"`
}

func (e *Executor) findNotes(ctx context.Context, userID int64, argsJSON string) (string, error) {
	var args findNotesArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse find_notes args: %w", err)
	}

	params := repo_notes.FindParams{
		UserID: userID,
	}

	if args.Search != "" {
		params.Search = &args.Search
	}

	if args.Workspace != "" {
		ws, err := e.services.Workspaces().FindOrCreateByName(ctx, userID, args.Workspace)
		if err == nil {
			params.WorkspaceID = &ws.ID
		}
	}

	pageSize := uint32(50)
	params.PageSize = &pageSize

	result, err := e.services.Notes().Find(ctx, params)
	if err != nil {
		return "", fmt.Errorf("find notes: %w", err)
	}

	items := make([]map[string]any, 0, len(result.Items))
	for _, note := range result.Items {
		item := map[string]any{
			"id":    note.ID,
			"title": note.Title,
		}
		if note.Body != "" {
			// Truncate body for LLM context
			body := note.Body
			if len(body) > 500 {
				body = body[:500] + "..."
			}
			item["body"] = body
		}
		items = append(items, item)
	}

	return toJSON(map[string]any{
		"total_count": result.Count,
		"items":       items,
	})
}

type completeTodoArgs struct {
	ID string `json:"id"`
}

func (e *Executor) completeTodo(ctx context.Context, userID int64, argsJSON string) (string, error) {
	var args completeTodoArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse complete_todo args: %w", err)
	}

	todo, err := e.services.Todos().Complete(ctx, userID, args.ID)
	if err != nil {
		return "", fmt.Errorf("complete todo: %w", err)
	}

	return toJSON(map[string]any{
		"id":     todo.ID,
		"title":  todo.Title,
		"status": string(todo.Status),
	})
}

type updateTodoArgs struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	PlannedDate string `json:"planned_date"`
	Status      string `json:"status"`
	Workspace   string `json:"workspace"`
}

func (e *Executor) updateTodo(ctx context.Context, userID int64, argsJSON string) (string, error) {
	var args updateTodoArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse update_todo args: %w", err)
	}

	params := todos.UpdateParams{}

	if args.Title != "" {
		params.Title = &args.Title
	}
	if args.Description != "" {
		params.Description = &args.Description
	}
	if args.PlannedDate != "" {
		t := carbon.Parse(args.PlannedDate).StdTime()
		if !t.IsZero() {
			params.PlannedDate = &t
		}
	}
	if args.Status != "" {
		status := models.TodoStatusType(args.Status)
		params.Status = &status
	}
	if args.Workspace != "" {
		ws, err := e.services.Workspaces().FindOrCreateByName(ctx, userID, args.Workspace)
		if err == nil {
			params.WorkspaceID = &ws.ID
		}
	}

	todo, err := e.services.Todos().Update(ctx, userID, args.ID, params)
	if err != nil {
		return "", fmt.Errorf("update todo: %w", err)
	}

	return toJSON(map[string]any{
		"id":           todo.ID,
		"title":        todo.Title,
		"status":       string(todo.Status),
		"planned_date": todo.PlannedDate.Format("2006-01-02"),
	})
}

func (e *Executor) listWorkspaces(ctx context.Context, userID int64) (string, error) {
	params := repo_workspaces.FindParams{
		UserID: userID,
	}

	pageSize := uint32(50)
	params.PageSize = &pageSize

	result, err := e.services.Workspaces().Find(ctx, params)
	if err != nil {
		return "", fmt.Errorf("list workspaces: %w", err)
	}

	items := make([]map[string]any, 0, len(result.Items))
	for _, ws := range result.Items {
		item := map[string]any{
			"id":   ws.ID,
			"name": ws.Name,
		}
		if ws.Description != "" {
			item["description"] = ws.Description
		}
		if ws.Archived {
			item["archived"] = true
		}
		items = append(items, item)
	}

	return toJSON(map[string]any{
		"total_count": result.Count,
		"items":       items,
	})
}

type saveToInboxArgs struct {
	Text string `json:"text"`
}

func (e *Executor) saveToInbox(ctx context.Context, userID int64, argsJSON string) (string, error) {
	var args saveToInboxArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse save_to_inbox args: %w", err)
	}

	item, err := e.services.Inbox().Create(ctx, args.Text, strconv.FormatInt(userID, 10))
	if err != nil {
		return "", fmt.Errorf("save to inbox: %w", err)
	}

	return toJSON(map[string]any{
		"id":   item.ID,
		"text": item.Data,
	})
}

// --- New tools: delete, update_note, stats, inbox ---

type deleteTodoArgs struct {
	ID string `json:"id"`
}

func (e *Executor) deleteTodo(ctx context.Context, userID int64, argsJSON string) (string, error) {
	var args deleteTodoArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse delete_todo args: %w", err)
	}

	if err := e.services.Todos().Delete(ctx, userID, args.ID); err != nil {
		return "", fmt.Errorf("delete todo: %w", err)
	}

	return toJSON(map[string]any{
		"deleted": true,
		"id":      args.ID,
	})
}

type deleteNoteArgs struct {
	ID string `json:"id"`
}

func (e *Executor) deleteNote(ctx context.Context, userID int64, argsJSON string) (string, error) {
	var args deleteNoteArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse delete_note args: %w", err)
	}

	if err := e.services.Notes().Delete(ctx, userID, args.ID); err != nil {
		return "", fmt.Errorf("delete note: %w", err)
	}

	return toJSON(map[string]any{
		"deleted": true,
		"id":      args.ID,
	})
}

type updateNoteArgs struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Workspace string `json:"workspace"`
}

func (e *Executor) updateNote(ctx context.Context, userID int64, argsJSON string) (string, error) {
	var args updateNoteArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse update_note args: %w", err)
	}

	params := notes.UpdateParams{}

	if args.Title != "" {
		params.Title = &args.Title
	}
	if args.Body != "" {
		params.Body = &args.Body
	}
	if args.Workspace != "" {
		ws, err := e.services.Workspaces().FindOrCreateByName(ctx, userID, args.Workspace)
		if err == nil {
			params.WorkspaceID = &ws.ID
		}
	}

	note, err := e.services.Notes().Update(ctx, userID, args.ID, params)
	if err != nil {
		return "", fmt.Errorf("update note: %w", err)
	}

	return toJSON(map[string]any{
		"id":    note.ID,
		"title": note.Title,
	})
}

type getTodoStatsArgs struct {
	DateFrom  string `json:"date_from"`
	DateTo    string `json:"date_to"`
	Workspace string `json:"workspace"`
}

func (e *Executor) getTodoStats(ctx context.Context, userID int64, argsJSON string) (string, error) {
	var args getTodoStatsArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse get_todo_stats args: %w", err)
	}

	baseParams := repo_todos.FindParams{
		UserID: userID,
	}

	if args.DateFrom != "" {
		t := carbon.Parse(args.DateFrom).StartOfDay().StdTime()
		if !t.IsZero() {
			baseParams.DateFrom = &t
		}
	}
	if args.DateTo != "" {
		t := carbon.Parse(args.DateTo).EndOfDay().StdTime()
		if !t.IsZero() {
			baseParams.DateTo = &t
		}
	}
	if args.Workspace != "" {
		ws, err := e.services.Workspaces().FindOrCreateByName(ctx, userID, args.Workspace)
		if err == nil {
			baseParams.WorkspaceID = &ws.ID
		}
	}

	// Count pending
	pendingParams := baseParams
	pendingParams.Statuses = []models.TodoStatusType{models.TodoStatusPending}
	pendingCount, _ := e.services.Todos().Count(ctx, pendingParams)

	// Count in progress
	inProgressParams := baseParams
	inProgressParams.Statuses = []models.TodoStatusType{models.TodoStatusInProgress}
	inProgressCount, _ := e.services.Todos().Count(ctx, inProgressParams)

	// Count completed
	completedParams := baseParams
	completedParams.Statuses = []models.TodoStatusType{models.TodoStatusCompleted}
	completedCount, _ := e.services.Todos().Count(ctx, completedParams)

	// Count cancelled
	cancelledParams := baseParams
	cancelledParams.Statuses = []models.TodoStatusType{models.TodoStatusCancelled}
	cancelledCount, _ := e.services.Todos().Count(ctx, cancelledParams)

	total := pendingCount + inProgressCount + completedCount + cancelledCount

	return toJSON(map[string]any{
		"total":       total,
		"pending":     pendingCount,
		"in_progress": inProgressCount,
		"completed":   completedCount,
		"cancelled":   cancelledCount,
	})
}

func (e *Executor) listInbox(ctx context.Context, userID int64) (string, error) {
	pageSize := uint32(50)
	result, err := e.services.Inbox().List(ctx, strconv.FormatInt(userID, 10), nil, &pageSize)
	if err != nil {
		return "", fmt.Errorf("list inbox: %w", err)
	}

	items := make([]map[string]any, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, map[string]any{
			"id":         item.ID,
			"text":       item.Data,
			"created_at": item.CreatedAt.Format("2006-01-02 15:04"),
		})
	}

	return toJSON(map[string]any{
		"total_count": result.Count,
		"items":       items,
	})
}

type convertInboxArgs struct {
	InboxID     string `json:"inbox_id"`
	ConvertTo   string `json:"convert_to"`
	Title       string `json:"title"`
	PlannedDate string `json:"planned_date"`
	Workspace   string `json:"workspace"`
}

func (e *Executor) convertInbox(ctx context.Context, userID int64, argsJSON string) (string, error) {
	var args convertInboxArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse convert_inbox args: %w", err)
	}

	var workspaceID *string
	if args.Workspace != "" {
		ws, err := e.services.Workspaces().FindOrCreateByName(ctx, userID, args.Workspace)
		if err == nil {
			workspaceID = &ws.ID
		}
	}

	switch args.ConvertTo {
	case "todo":
		plannedDate := time.Now()
		if args.PlannedDate != "" {
			parsed := carbon.Parse(args.PlannedDate).StdTime()
			if !parsed.IsZero() {
				plannedDate = parsed
			}
		}

		todoID, err := e.services.Inbox().ConvertToTodo(ctx, args.InboxID, userID, args.Title, "", plannedDate, workspaceID)
		if err != nil {
			return "", fmt.Errorf("convert inbox to todo: %w", err)
		}
		return toJSON(map[string]any{
			"converted_to": "todo",
			"new_id":       todoID,
		})

	case "note":
		noteID, err := e.services.Inbox().ConvertToNote(ctx, args.InboxID, userID, args.Title, "", workspaceID)
		if err != nil {
			return "", fmt.Errorf("convert inbox to note: %w", err)
		}
		return toJSON(map[string]any{
			"converted_to": "note",
			"new_id":       noteID,
		})

	default:
		return "", fmt.Errorf("invalid convert_to value: %s (must be 'todo' or 'note')", args.ConvertTo)
	}
}

func toJSON(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

