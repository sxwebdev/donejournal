package api

import (
	"database/sql"

	authv1 "github.com/sxwebdev/donejournal/api/gen/go/donejournal/auth/v1"
	inboxv1 "github.com/sxwebdev/donejournal/api/gen/go/donejournal/inbox/v1"
	notesv1 "github.com/sxwebdev/donejournal/api/gen/go/donejournal/notes/v1"
	tagsv1 "github.com/sxwebdev/donejournal/api/gen/go/donejournal/tags/v1"
	todosv1 "github.com/sxwebdev/donejournal/api/gen/go/donejournal/todos/v1"
	workspacesv1 "github.com/sxwebdev/donejournal/api/gen/go/donejournal/workspaces/v1"
	"github.com/sxwebdev/donejournal/internal/models"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_workspaces"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func todoToProto(t *models.Todo) *todosv1.Todo {
	pb := &todosv1.Todo{
		Id:                 t.ID,
		Title:              t.Title,
		Description:        t.Description,
		Status:             todoStatusToProto(t.Status),
		Priority:           todoPriorityToProto(t.Priority),
		PlannedDate:        timestamppb.New(t.PlannedDate),
		CreatedAt:          timestamppb.New(t.CreatedAt),
		UpdatedAt:          timestamppb.New(t.UpdatedAt),
		WorkspaceId:        nullStringPtr(t.WorkspaceID),
		RecurrenceRule:     nullStringPtr(t.RecurrenceRule),
		RecurrenceParentId: nullStringPtr(t.RecurrenceParentID),
	}
	if t.CompletedAt.Valid {
		pb.CompletedAt = timestamppb.New(t.CompletedAt.Time)
	}
	return pb
}

func todoStatusToProto(s models.TodoStatusType) todosv1.TodoStatus {
	switch s {
	case models.TodoStatusPending:
		return todosv1.TodoStatus_TODO_STATUS_PENDING
	case models.TodoStatusInProgress:
		return todosv1.TodoStatus_TODO_STATUS_IN_PROGRESS
	case models.TodoStatusCompleted:
		return todosv1.TodoStatus_TODO_STATUS_COMPLETED
	case models.TodoStatusCancelled:
		return todosv1.TodoStatus_TODO_STATUS_CANCELLED
	default:
		return todosv1.TodoStatus_TODO_STATUS_UNSPECIFIED
	}
}

func todoStatusFromProto(s todosv1.TodoStatus) models.TodoStatusType {
	switch s {
	case todosv1.TodoStatus_TODO_STATUS_PENDING:
		return models.TodoStatusPending
	case todosv1.TodoStatus_TODO_STATUS_IN_PROGRESS:
		return models.TodoStatusInProgress
	case todosv1.TodoStatus_TODO_STATUS_COMPLETED:
		return models.TodoStatusCompleted
	case todosv1.TodoStatus_TODO_STATUS_CANCELLED:
		return models.TodoStatusCancelled
	default:
		return models.TodoStatusPending
	}
}

func todoPriorityToProto(p models.TodoPriorityType) todosv1.TodoPriority {
	switch p {
	case models.TodoPriorityNone:
		return todosv1.TodoPriority_TODO_PRIORITY_NONE
	case models.TodoPriorityLow:
		return todosv1.TodoPriority_TODO_PRIORITY_LOW
	case models.TodoPriorityMedium:
		return todosv1.TodoPriority_TODO_PRIORITY_MEDIUM
	case models.TodoPriorityHigh:
		return todosv1.TodoPriority_TODO_PRIORITY_HIGH
	case models.TodoPriorityCritical:
		return todosv1.TodoPriority_TODO_PRIORITY_CRITICAL
	default:
		return todosv1.TodoPriority_TODO_PRIORITY_UNSPECIFIED
	}
}

func todoPriorityFromProto(p todosv1.TodoPriority) models.TodoPriorityType {
	switch p {
	case todosv1.TodoPriority_TODO_PRIORITY_NONE:
		return models.TodoPriorityNone
	case todosv1.TodoPriority_TODO_PRIORITY_LOW:
		return models.TodoPriorityLow
	case todosv1.TodoPriority_TODO_PRIORITY_MEDIUM:
		return models.TodoPriorityMedium
	case todosv1.TodoPriority_TODO_PRIORITY_HIGH:
		return models.TodoPriorityHigh
	case todosv1.TodoPriority_TODO_PRIORITY_CRITICAL:
		return models.TodoPriorityCritical
	default:
		return models.TodoPriorityNone
	}
}

func inboxItemToProto(i *models.Inbox) *inboxv1.InboxItem {
	return &inboxv1.InboxItem{
		Id:             i.ID,
		Data:           i.Data,
		AdditionalData: string(i.AdditionalData),
		CreatedAt:      timestamppb.New(i.CreatedAt),
		UpdatedAt:      timestamppb.New(i.UpdatedAt),
	}
}

func noteToProto(n *models.Note) *notesv1.Note {
	return &notesv1.Note{
		Id:          n.ID,
		Title:       n.Title,
		Body:        n.Body,
		CreatedAt:   timestamppb.New(n.CreatedAt),
		UpdatedAt:   timestamppb.New(n.UpdatedAt),
		WorkspaceId: nullStringPtr(n.WorkspaceID),
	}
}

func workspaceToProto(w *models.Workspace) *workspacesv1.Workspace {
	return &workspacesv1.Workspace{
		Id:          w.ID,
		Name:        w.Name,
		Description: w.Description,
		Archived:    w.Archived,
		CreatedAt:   timestamppb.New(w.CreatedAt),
		UpdatedAt:   timestamppb.New(w.UpdatedAt),
	}
}

func workspaceStatsToProto(s *repo_workspaces.WorkspaceStats) *workspacesv1.WorkspaceStats {
	return &workspacesv1.WorkspaceStats{
		Workspace:          workspaceToProto(&s.Workspace),
		TodoCount:          s.TodoCount,
		NoteCount:          s.NoteCount,
		CompletedTodoCount: s.CompletedTodoCount,
	}
}

func tokenDataToUserProto(data TokenData) *authv1.User {
	return &authv1.User{
		Id:        data.TelegramID,
		FirstName: data.FirstName,
		LastName:  data.LastName,
		Username:  data.Username,
		PhotoUrl:  data.PhotoURL,
	}
}

func tagToProto(t *models.Tag) *tagsv1.Tag {
	return &tagsv1.Tag{
		Id:        t.ID,
		Name:      t.Name,
		Color:     t.Color,
		CreatedAt: timestamppb.New(t.CreatedAt),
		UpdatedAt: timestamppb.New(t.UpdatedAt),
	}
}

func nullStringPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}
