package domain

import (
	"time"

	"github.com/google/uuid"
)

// Project represents a monitored application.
type Project struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	OwnerID   uuid.UUID `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

// APIKey represents an authentication key scoped to a project.
type APIKey struct {
	ID        uuid.UUID  `json:"id"`
	KeyHash   string     `json:"-"`
	ProjectID uuid.UUID  `json:"project_id"`
	Label     string     `json:"label"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}
