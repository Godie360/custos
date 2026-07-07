package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/iPFSoftwares/custos/internal/domain"
)

// ProjectsStore is the minimal interface needed by ProjectsHandler.
type ProjectsStore interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error)
}

// APIKeysStore is the minimal interface for API key operations.
type APIKeysStore interface {
	Create(ctx context.Context, key *domain.APIKey) error
	Revoke(ctx context.Context, id uuid.UUID) error
}

// ProjectsHandler handles project-related HTTP endpoints.
type ProjectsHandler struct {
	projects ProjectsStore
	apiKeys  APIKeysStore
}

// NewProjectsHandler creates a ProjectsHandler.
func NewProjectsHandler(projects ProjectsStore, apiKeys APIKeysStore) *ProjectsHandler {
	return &ProjectsHandler{projects: projects, apiKeys: apiKeys}
}

// List handles GET /api/v1/projects.
// In a full implementation this would filter by authenticated user; here we
// return a not-implemented stub so the route compiles and is routable.
func (h *ProjectsHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]any{})
}

// CreateAPIKey handles POST /api/v1/projects/{id}/keys.
func (h *ProjectsHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	if _, err := h.projects.GetByID(r.Context(), projectID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "project not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to retrieve project", http.StatusInternalServerError)
		return
	}

	var req struct {
		Label string `json:"label"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	key := &domain.APIKey{
		ID:        uuid.New(),
		ProjectID: projectID,
		Label:     req.Label,
	}

	if err := h.apiKeys.Create(r.Context(), key); err != nil {
		http.Error(w, "failed to create API key", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(key)
}

// RevokeAPIKey handles DELETE /api/v1/projects/{id}/keys/{kid}.
func (h *ProjectsHandler) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	keyID, err := uuid.Parse(chi.URLParam(r, "kid"))
	if err != nil {
		http.Error(w, "invalid key id", http.StatusBadRequest)
		return
	}

	if err := h.apiKeys.Revoke(r.Context(), keyID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "key not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to revoke key", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
