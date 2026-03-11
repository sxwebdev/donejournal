package api

import (
	"time"

	authv1 "github.com/sxwebdev/donejournal/api/gen/go/donejournal/auth/v1"
	inboxv1 "github.com/sxwebdev/donejournal/api/gen/go/donejournal/inbox/v1"
	notesv1 "github.com/sxwebdev/donejournal/api/gen/go/donejournal/notes/v1"
	todosv1 "github.com/sxwebdev/donejournal/api/gen/go/donejournal/todos/v1"
	"github.com/sxwebdev/donejournal/internal/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func todoToProto(t *models.Todo) *todosv1.Todo {
	pb := &todosv1.Todo{
		Id:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Status:      todoStatusToProto(t.Status),
		PlannedDate: timestamppb.New(t.PlannedDate),
		CreatedAt:   timestamppb.New(t.CreatedAt),
		UpdatedAt:   timestamppb.New(t.UpdatedAt),
	}
	if t.CompletedAt != nil {
		pb.CompletedAt = timestamppb.New(*t.CompletedAt)
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
		Id:        n.ID,
		Title:     n.Title,
		Body:      n.Body,
		CreatedAt: timestamppb.New(n.CreatedAt),
		UpdatedAt: timestamppb.New(n.UpdatedAt),
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

func timeToOptionalTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}
