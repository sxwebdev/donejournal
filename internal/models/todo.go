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

type TodoPriorityType string

const (
	TodoPriorityNone     TodoPriorityType = "none"
	TodoPriorityLow      TodoPriorityType = "low"
	TodoPriorityMedium   TodoPriorityType = "medium"
	TodoPriorityHigh     TodoPriorityType = "high"
	TodoPriorityCritical TodoPriorityType = "critical"
)

// Validate validates the TodoPriorityType
func (tpt TodoPriorityType) Validate() error {
	switch tpt {
	case TodoPriorityNone, TodoPriorityLow, TodoPriorityMedium, TodoPriorityHigh, TodoPriorityCritical:
		return nil
	default:
		return fmt.Errorf("invalid priority: %s", tpt)
	}
}
