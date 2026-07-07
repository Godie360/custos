package domain

import (
	"time"

	"github.com/google/uuid"
)

// Project represents a monitored application.
type Project struct {
	ID        uuid.UUID
	Name      string
	Slug      string
	OwnerID   uuid.UUID
	CreatedAt time.Time
}

// APIKey represents an authentication key scoped to a project.
type APIKey struct {
	ID        uuid.UUID
	KeyHash   string
	ProjectID uuid.UUID
	Label     string
	CreatedAt time.Time
	ExpiresAt *time.Time
	RevokedAt *time.Time
}
