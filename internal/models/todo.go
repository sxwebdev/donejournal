package models

type TodoStatusType string

const (
	TodoStatusPending    TodoStatusType = "pending"
	TodoStatusInProgress TodoStatusType = "inprogress"
	TodoStatusCompleted  TodoStatusType = "completed"
	TodoStatusCancelled  TodoStatusType = "cancelled"
)
