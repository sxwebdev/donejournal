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
							"type":        "string",
							"description": "Optional longer description",
						},
						"planned_date": map[string]any{
							"type":        "string",
							"description": "Date in YYYY-MM-DD format. Defaults to today if not specified.",
						},
						"status": map[string]any{
							"type":        "string",
							"enum":        []string{"pending", "completed"},
							"description": "pending = planned task, completed = already done. Defaults to pending.",
						},
						"workspace": map[string]any{
							"type":        "string",
							"description": "Optional workspace/project name. Will be auto-created if doesn't exist.",
						},
					},
					"required": []string{"title"},
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
							"type":        "string",
							"description": "Full note content, supports markdown",
						},
						"workspace": map[string]any{
							"type":        "string",
							"description": "Optional workspace/project name",
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
							"type":        "array",
							"items":       map[string]any{"type": "string", "enum": []string{"pending", "inprogress", "completed", "cancelled"}},
							"description": "Filter by status. If not specified, returns pending and inprogress.",
						},
						"date_from": map[string]any{
							"type":        "string",
							"description": "Start date filter in YYYY-MM-DD format",
						},
						"date_to": map[string]any{
							"type":        "string",
							"description": "End date filter in YYYY-MM-DD format",
						},
						"workspace": map[string]any{
							"type":        "string",
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
							"type":        "string",
							"description": "Text to search in note title and body",
						},
						"workspace": map[string]any{
							"type":        "string",
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
							"type":        "string",
							"description": "New title",
						},
						"description": map[string]any{
							"type":        "string",
							"description": "New description",
						},
						"planned_date": map[string]any{
							"type":        "string",
							"description": "New date in YYYY-MM-DD format",
						},
						"status": map[string]any{
							"type":        "string",
							"enum":        []string{"pending", "inprogress", "completed", "cancelled"},
							"description": "New status",
						},
						"workspace": map[string]any{
							"type":        "string",
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
							"type":        "string",
							"description": "New title",
						},
						"body": map[string]any{
							"type":        "string",
							"description": "New body content (markdown)",
						},
						"workspace": map[string]any{
							"type":        "string",
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
							"type":        "string",
							"description": "Start date in YYYY-MM-DD format",
						},
						"date_to": map[string]any{
							"type":        "string",
							"description": "End date in YYYY-MM-DD format",
						},
						"workspace": map[string]any{
							"type":        "string",
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
							"type":        "string",
							"description": "Title for the new todo/note. If empty, uses inbox item text.",
						},
						"planned_date": map[string]any{
							"type":        "string",
							"description": "Date in YYYY-MM-DD (only for todo conversion). Defaults to today.",
						},
						"workspace": map[string]any{
							"type":        "string",
							"description": "Optional workspace name",
						},
					},
					"required": []string{"inbox_id", "convert_to"},
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
