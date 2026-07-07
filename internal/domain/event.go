package domain

import (
	"time"

	"github.com/google/uuid"
)

// RawEvent represents an error event received from a language SDK.
type RawEvent struct {
	ID                uuid.UUID  `json:"id"`
	IssueID           *uuid.UUID `json:"issue_id,omitempty"`
	ProjectID         uuid.UUID  `json:"project_id"`
	Service           string     `json:"service"`
	Environment       string     `json:"environment"`
	ErrorType         string     `json:"error_type"`
	Message           string     `json:"message"`
	StackTrace        []string   `json:"stack_trace,omitempty"`
	RawBody           string     `json:"-"`
	RedactedBody      string     `json:"-"`
	ReceivedAt        time.Time  `json:"received_at"`
	RetentionDeleteAt *time.Time `json:"-"`
}
