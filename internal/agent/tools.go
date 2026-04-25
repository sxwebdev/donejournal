package agent

import (
	"encoding/json"

	"github.com/sxwebdev/donejournal/internal/agent/provider"
)

func toolDefinitions() []provider.ToolDefinition {
	return []provider.ToolDefinition{
		{
			Type: "function",
			Function: provider.FunctionDef{
				Name:        "create_todo",
				Description: "Create a new task/todo. Use for planned tasks or completed tasks (done items).",
				Parameters: mustJSON(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"title": map[string]any{
							"type":        "string",
							"description": "Task title",
						},
						"description": map[string]any{
							"type":        []string{"string", "null"},
							"description": "Optional longer description",
						},
						"planned_date": map[string]any{
							"type":        []string{"string", "null"},
							"description": "Date in YYYY-MM-DD format. Defaults to today if not specified.",
						},
						"status": map[string]any{
							"type":        []string{"string", "null"},
							"enum":        []string{"pending", "completed"},
							"description": "pending = planned task, completed = already done. Defaults to pending.",
						},
						"workspace": map[string]any{
							"type":        []string{"string", "null"},
							"description": "Optional workspace/project name. Will be auto-created if doesn't exist.",
						},
						"priority": map[string]any{
							"type":        []string{"string", "null"},
							"enum":        []string{"none", "low", "medium", "high", "critical"},
							"description": "Task priority level. Default: none.",
						},
						"tags": map[string]any{
							"type":        "array",
							"items":       map[string]any{"type": "string"},
							"description": "Optional tag names to attach. Tags are auto-created if they don't exist.",
						},
					},
					"required": []string{"title"},
				}),
			},
		},
		{
			Type: "function",
			Function: provider.FunctionDef{
				Name:        "create_recurring_todo",
				Description: "Create a recurring task that automatically repeats. Use when user says 'every day', 'weekly', 'каждый день', 'еженедельно', 'каждый понедельник', etc.",
				Parameters: mustJSON(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"title": map[string]any{
							"type":        "string",
							"description": "Task title",
						},
						"description": map[string]any{
							"type":        []string{"string", "null"},
							"description": "Optional longer description",
						},
						"planned_date": map[string]any{
							"type":        []string{"string", "null"},
							"description": "First occurrence date in YYYY-MM-DD format. Defaults to today.",
						},
						"recurrence_rule": map[string]any{
							"type":        "string",
							"enum":        []string{"daily", "weekly", "monthly"},
							"description": "How often to repeat: daily (every day), weekly (every 7 days), monthly (every month).",
						},
						"workspace": map[string]any{
							"type":        []string{"string", "null"},
							"description": "Optional workspace/project name.",
						},
						"priority": map[string]any{
							"type":        []string{"string", "null"},
							"enum":        []string{"none", "low", "medium", "high", "critical"},
							"description": "Task priority level. Default: none.",
						},
						"tags": map[string]any{
							"type":        "array",
							"items":       map[string]any{"type": "string"},
							"description": "Optional tag names to attach.",
						},
					},
					"required": []string{"title", "recurrence_rule"},
				}),
			},
		},
		{
			Type: "function",
			Function: provider.FunctionDef{
				Name:        "create_note",
				Description: "Create a note. Use for ideas, reference material, things to remember.",
				Parameters: mustJSON(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"title": map[string]any{
							"type":        "string",
							"description": "Note title",
						},
						"body": map[string]any{
							"type":        []string{"string", "null"},
							"description": "Full note content, supports markdown",
						},
						"workspace": map[string]any{
							"type":        []string{"string", "null"},
							"description": "Optional workspace/project name",
						},
						"tags": map[string]any{
							"type":        "array",
							"items":       map[string]any{"type": "string"},
							"description": "Optional tag names to attach. Tags are auto-created if they don't exist.",
						},
					},
					"required": []string{"title"},
				}),
			},
		},
		{
			Type: "function",
			Function: provider.FunctionDef{
				Name:        "find_todos",
				Description: "Search and list user's todos/tasks with filters. Returns matching tasks.",
				Parameters: mustJSON(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"status": map[string]any{
							"type":        []string{"array", "null"},
							"items":       map[string]any{"type": "string", "enum": []string{"pending", "inprogress", "completed", "cancelled"}},
							"description": "Filter by status. Pass as a JSON array, e.g. [\"completed\"]. If not specified, returns pending and inprogress.",
						},
						"date_from": map[string]any{
							"type":        []string{"string", "null"},
							"description": "Start date filter (YYYY-MM-DD), inclusive. When status=[\"completed\"] this filters by completed_at (when the task was finished); otherwise by planned_date.",
						},
						"date_to": map[string]any{
							"type":        []string{"string", "null"},
							"description": "End date filter (YYYY-MM-DD), inclusive. When status=[\"completed\"] this filters by completed_at; otherwise by planned_date.",
						},
						"workspace": map[string]any{
							"type":        []string{"string", "null"},
							"description": "Filter by workspace name",
						},
					},
				}),
			},
		},
		{
			Type: "function",
			Function: provider.FunctionDef{
				Name:        "find_notes",
				Description: "Search user's notes by text query.",
				Parameters: mustJSON(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"search": map[string]any{
							"type":        []string{"string", "null"},
							"description": "Text to search in note title and body",
						},
						"workspace": map[string]any{
							"type":        []string{"string", "null"},
							"description": "Filter by workspace name",
						},
					},
				}),
			},
		},
		{
			Type: "function",
			Function: provider.FunctionDef{
				Name:        "complete_todo",
				Description: "Mark a todo as completed.",
				Parameters: mustJSON(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id": map[string]any{
							"type":        "string",
							"description": "Todo ID to complete",
						},
					},
					"required": []string{"id"},
				}),
			},
		},
		{
			Type: "function",
			Function: provider.FunctionDef{
				Name:        "update_todo",
				Description: "Update an existing todo. Can change title, description, date, status, workspace.",
				Parameters: mustJSON(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id": map[string]any{
							"type":        "string",
							"description": "Todo ID to update",
						},
						"title": map[string]any{
							"type":        []string{"string", "null"},
							"description": "New title",
						},
						"description": map[string]any{
							"type":        []string{"string", "null"},
							"description": "New description",
						},
						"planned_date": map[string]any{
							"type":        []string{"string", "null"},
							"description": "New date in YYYY-MM-DD format",
						},
						"status": map[string]any{
							"type":        []string{"string", "null"},
							"enum":        []string{"pending", "inprogress", "completed", "cancelled"},
							"description": "New status",
						},
						"workspace": map[string]any{
							"type":        []string{"string", "null"},
							"description": "New workspace name",
						},
						"priority": map[string]any{
							"type":        []string{"string", "null"},
							"enum":        []string{"none", "low", "medium", "high", "critical"},
							"description": "New priority level",
						},
					},
					"required": []string{"id"},
				}),
			},
		},
		{
			Type: "function",
			Function: provider.FunctionDef{
				Name:        "list_workspaces",
				Description: "List all user's workspaces/projects.",
				Parameters: mustJSON(map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				}),
			},
		},
		{
			Type: "function",
			Function: provider.FunctionDef{
				Name:        "save_to_inbox",
				Description: "Save unstructured text to inbox for later processing. Use when intent is unclear or user explicitly asks to save to inbox.",
				Parameters: mustJSON(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"text": map[string]any{
							"type":        "string",
							"description": "Text to save",
						},
					},
					"required": []string{"text"},
				}),
			},
		},
		{
			Type: "function",
			Function: provider.FunctionDef{
				Name:        "delete_todo",
				Description: "Delete a todo/task permanently. IMPORTANT: Always confirm with the user before deleting.",
				Parameters: mustJSON(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id": map[string]any{
							"type":        "string",
							"description": "Todo ID to delete",
						},
					},
					"required": []string{"id"},
				}),
			},
		},
		{
			Type: "function",
			Function: provider.FunctionDef{
				Name:        "bulk_delete_todos",
				Description: "Delete multiple todos matching filters in one call. Use for queries like 'delete all completed tasks before DATE', 'delete cancelled tasks in workspace X'. REQUIRED two-step flow: (1) call find_todos with the same filters, show the user what will be deleted, ask for confirmation; (2) only after explicit user confirmation, call this with confirmed=true. At least one filter (status / date_from / date_to / workspace / tags) is required. Do NOT loop delete_todo for bulk operations.",
				Parameters: mustJSON(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"status": map[string]any{
							"type":        []string{"array", "null"},
							"items":       map[string]any{"type": "string", "enum": []string{"pending", "inprogress", "completed", "cancelled"}},
							"description": "Filter by status. Pass as a JSON array, e.g. [\"completed\"]. When status=[\"completed\"], date_from/date_to filter by completed_at; otherwise by planned_date.",
						},
						"date_from": map[string]any{
							"type":        []string{"string", "null"},
							"description": "Start date filter (YYYY-MM-DD), inclusive. Same field semantics as find_todos.",
						},
						"date_to": map[string]any{
							"type":        []string{"string", "null"},
							"description": "End date filter (YYYY-MM-DD), inclusive. Same field semantics as find_todos.",
						},
						"workspace": map[string]any{
							"type":        []string{"string", "null"},
							"description": "Filter by workspace name",
						},
						"tags": map[string]any{
							"type":        []string{"array", "null"},
							"items":       map[string]any{"type": "string"},
							"description": "Filter by tag names (matches todos that have ANY of these tags). Pass as a JSON array, e.g. [\"work\"].",
						},
						"confirmed": map[string]any{
							"type":        "boolean",
							"description": "Must be true. The tool refuses to delete without explicit user confirmation. Set true ONLY after the user has confirmed the deletion in this conversation.",
						},
					},
					"required": []string{"confirmed"},
				}),
			},
		},
		{
			Type: "function",
			Function: provider.FunctionDef{
				Name:        "delete_note",
				Description: "Delete a note permanently. IMPORTANT: Always confirm with the user before deleting.",
				Parameters: mustJSON(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id": map[string]any{
							"type":        "string",
							"description": "Note ID to delete",
						},
					},
					"required": []string{"id"},
				}),
			},
		},
		{
			Type: "function",
			Function: provider.FunctionDef{
				Name:        "update_note",
				Description: "Update an existing note. Can change title, body, workspace.",
				Parameters: mustJSON(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id": map[string]any{
							"type":        "string",
							"description": "Note ID to update",
						},
						"title": map[string]any{
							"type":        []string{"string", "null"},
							"description": "New title",
						},
						"body": map[string]any{
							"type":        []string{"string", "null"},
							"description": "New body content (markdown)",
						},
						"workspace": map[string]any{
							"type":        []string{"string", "null"},
							"description": "New workspace name",
						},
					},
					"required": []string{"id"},
				}),
			},
		},
		{
			Type: "function",
			Function: provider.FunctionDef{
				Name:        "get_todo_stats",
				Description: "Get statistics about user's todos: count of pending, completed, etc. for a date range.",
				Parameters: mustJSON(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"date_from": map[string]any{
							"type":        []string{"string", "null"},
							"description": "Start date in YYYY-MM-DD format",
						},
						"date_to": map[string]any{
							"type":        []string{"string", "null"},
							"description": "End date in YYYY-MM-DD format",
						},
						"workspace": map[string]any{
							"type":        []string{"string", "null"},
							"description": "Filter by workspace name",
						},
					},
				}),
			},
		},
		{
			Type: "function",
			Function: provider.FunctionDef{
				Name:        "list_inbox",
				Description: "List items in user's inbox (unprocessed messages saved for later).",
				Parameters: mustJSON(map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				}),
			},
		},
		{
			Type: "function",
			Function: provider.FunctionDef{
				Name:        "convert_inbox",
				Description: "Convert an inbox item to a todo or note, then remove it from inbox.",
				Parameters: mustJSON(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"inbox_id": map[string]any{
							"type":        "string",
							"description": "Inbox item ID to convert",
						},
						"convert_to": map[string]any{
							"type":        "string",
							"enum":        []string{"todo", "note"},
							"description": "Convert to todo or note",
						},
						"title": map[string]any{
							"type":        []string{"string", "null"},
							"description": "Title for the new todo/note. If empty, uses inbox item text.",
						},
						"planned_date": map[string]any{
							"type":        []string{"string", "null"},
							"description": "Date in YYYY-MM-DD (only for todo conversion). Defaults to today.",
						},
						"workspace": map[string]any{
							"type":        []string{"string", "null"},
							"description": "Optional workspace name",
						},
					},
					"required": []string{"inbox_id", "convert_to"},
				}),
			},
		},
		{
			Type: "function",
			Function: provider.FunctionDef{
				Name:        "manage_tags",
				Description: "Create, list, or delete tags/labels. Tags are colored labels that can be attached to todos and notes.",
				Parameters: mustJSON(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"action": map[string]any{
							"type":        "string",
							"enum":        []string{"create", "list", "delete"},
							"description": "Action to perform",
						},
						"name": map[string]any{
							"type":        []string{"string", "null"},
							"description": "Tag name (required for create)",
						},
						"color": map[string]any{
							"type":        []string{"string", "null"},
							"description": "Tag color as hex (e.g. #ef4444). Defaults to #6366f1",
						},
						"id": map[string]any{
							"type":        []string{"string", "null"},
							"description": "Tag ID (required for delete)",
						},
					},
					"required": []string{"action"},
				}),
			},
		},
		{
			Type: "function",
			Function: provider.FunctionDef{
				Name:        "tag_entity",
				Description: "Add tags to a todo or note. Tags will be auto-created if they don't exist.",
				Parameters: mustJSON(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"entity_type": map[string]any{
							"type":        "string",
							"enum":        []string{"todo", "note"},
							"description": "Type of entity to tag",
						},
						"entity_id": map[string]any{
							"type":        "string",
							"description": "ID of the todo or note",
						},
						"tags": map[string]any{
							"type":        "array",
							"items":       map[string]any{"type": "string"},
							"description": "Tag names to add",
						},
					},
					"required": []string{"entity_type", "entity_id", "tags"},
				}),
			},
		},
		{
			Type: "function",
			Function: provider.FunctionDef{
				Name:        "find_by_tag",
				Description: "Find todos or notes that have specific tags.",
				Parameters: mustJSON(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"entity_type": map[string]any{
							"type":        "string",
							"enum":        []string{"todo", "note"},
							"description": "Type of entity to search",
						},
						"tags": map[string]any{
							"type":        "array",
							"items":       map[string]any{"type": "string"},
							"description": "Tag names to filter by",
						},
					},
					"required": []string{"entity_type", "tags"},
				}),
			},
		},
	}
}

func mustJSON(v any) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
