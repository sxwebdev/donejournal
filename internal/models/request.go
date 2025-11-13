package models

type RequestStatusType string

const (
	RequestStatusPending  RequestStatusType = "pending"
	RequestStatusApproved RequestStatusType = "approved"
	RequestStatusFailed   RequestStatusType = "failed"
)
