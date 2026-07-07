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
	ID                uuid.UUID
	Fingerprint       string
	ProjectID         uuid.UUID
	Service           string
	Environment       string
	FirstSeen         time.Time
	LastSeen          time.Time
	OccurrenceCount   int
	Status            IssueStatus
	Severity          string
	AIExplanation     string
	AILikelyCause     string
	AISuggestedChecks []string
}
