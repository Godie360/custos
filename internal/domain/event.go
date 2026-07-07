package domain

import (
	"time"

	"github.com/google/uuid"
)

// RawEvent represents an error event received from a language SDK.
type RawEvent struct {
	ID                uuid.UUID
	IssueID           *uuid.UUID
	ProjectID         uuid.UUID
	Service           string
	Environment       string
	ErrorType         string
	Message           string
	StackTrace        []string
	RawBody           string
	RedactedBody      string
	ReceivedAt        time.Time
	RetentionDeleteAt *time.Time
}
