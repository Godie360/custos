package domain

import (
	"time"

	"github.com/google/uuid"
)

// FilterField is the event field a rule matches against.
type FilterField string

const (
	FilterFieldErrorType   FilterField = "error_type"
	FilterFieldMessage     FilterField = "message"
	FilterFieldService     FilterField = "service"
	FilterFieldEnvironment FilterField = "environment"
)

// FilterOperator is the match strategy for a rule.
type FilterOperator string

const (
	FilterOperatorEquals     FilterOperator = "equals"
	FilterOperatorContains   FilterOperator = "contains"
	FilterOperatorStartsWith FilterOperator = "starts_with"
)

// FilterRule defines a condition that causes matching ingest events to be
// silently dropped before they are stored or analyzed.
type FilterRule struct {
	ID        uuid.UUID      `json:"id"`
	ProjectID uuid.UUID      `json:"project_id"`
	Field     FilterField    `json:"field"`
	Operator  FilterOperator `json:"operator"`
	Value     string         `json:"value"`
	CreatedAt time.Time      `json:"created_at"`
}

// Matches reports whether the rule matches the given raw event fields.
func (r *FilterRule) Matches(errorType, message, service, environment string) bool {
	var target string
	switch r.Field {
	case FilterFieldErrorType:
		target = errorType
	case FilterFieldMessage:
		target = message
	case FilterFieldService:
		target = service
	case FilterFieldEnvironment:
		target = environment
	default:
		return false
	}

	switch r.Operator {
	case FilterOperatorEquals:
		return target == r.Value
	case FilterOperatorContains:
		return contains(target, r.Value)
	case FilterOperatorStartsWith:
		return startsWith(target, r.Value)
	default:
		return false
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexStr(s, substr) >= 0
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func indexStr(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
