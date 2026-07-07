package domain

import (
	"time"

	"github.com/google/uuid"
)

// IssueStatus represents the lifecycle state of an issue.
type IssueStatus string

const (
	IssueStatusOpen           IssueStatus = "open"
	IssueStatusResolved       IssueStatus = "resolved"
	IssueStatusAnalysisFailed IssueStatus = "analysis_failed"
)

// Issue is a deduplicated, AI-enriched error group.
type Issue struct {
	ID                uuid.UUID   `json:"id"`
	Fingerprint       string      `json:"fingerprint"`
	ProjectID         uuid.UUID   `json:"project_id"`
	Service           string      `json:"service"`
	Environment       string      `json:"environment"`
	FirstSeen         time.Time   `json:"first_seen"`
	LastSeen          time.Time   `json:"last_seen"`
	OccurrenceCount   int         `json:"occurrence_count"`
	Status            IssueStatus `json:"status"`
	Severity          string      `json:"severity"`
	AIExplanation     string      `json:"ai_explanation,omitempty"`
	AILikelyCause     string      `json:"ai_likely_cause,omitempty"`
	AISuggestedChecks []string    `json:"ai_suggested_checks,omitempty"`
}
