package models

import "fmt"

type RequestStatusType string

const (
	RequestStatusPending   RequestStatusType = "pending"
	RequestStatusCompleted RequestStatusType = "completed"
	RequestStatusFailed    RequestStatusType = "failed"
)

// Validate validates the RequestStatusType
func (rst RequestStatusType) Validate() error {
	switch rst {
	case RequestStatusPending, RequestStatusCompleted, RequestStatusFailed:
		return nil
	default:
		return fmt.Errorf("invalid status: %s", rst)
	}
}
