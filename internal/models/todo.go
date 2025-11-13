package models

import "fmt"

type TodoStatusType string

const (
	TodoStatusPending    TodoStatusType = "pending"
	TodoStatusInProgress TodoStatusType = "inprogress"
	TodoStatusCompleted  TodoStatusType = "completed"
	TodoStatusCancelled  TodoStatusType = "cancelled"
)

// Validate validates the TodoStatusType
func (tst TodoStatusType) Validate() error {
	switch tst {
	case TodoStatusPending, TodoStatusInProgress, TodoStatusCompleted, TodoStatusCancelled:
		return nil
	default:
		return fmt.Errorf("invalid status: %s", tst)
	}
}
