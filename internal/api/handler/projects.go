package handler

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/iPFSoftwares/custos/internal/api/render"
	"github.com/iPFSoftwares/custos/internal/domain"
)

// systemOwnerID is the fixed seed user used as project owner before auth is implemented.
var systemOwnerID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

// ProjectsStore is the minimal interface needed by ProjectsHandler.
type ProjectsStore interface {
	Create(ctx context.Context, project *domain.Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error)
	List(ctx context.Context) ([]*domain.Project, error)
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

// Create handles POST /api/v1/projects.
func (h *ProjectsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Error(w, r, http.StatusBadRequest, "invalid_body", "request body must be valid JSON", err)
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		render.Error(w, r, http.StatusUnprocessableEntity, "missing_fields", "name is required", nil)
		return
	}

	project := &domain.Project{
		ID:        uuid.New(),
		Name:      strings.TrimSpace(req.Name),
		Slug:      slugify(req.Name),
		OwnerID:   systemOwnerID,
		CreatedAt: time.Now().UTC(),
	}

	if err := h.projects.Create(r.Context(), project); err != nil {
		if errors.Is(err, domain.ErrConflict) {
			render.Error(w, r, http.StatusConflict, "slug_conflict", "a project with this name already exists", nil)
			return
		}
		render.Error(w, r, http.StatusInternalServerError, "store_error", "failed to create project", err)
		return
	}

	render.JSON(w, http.StatusCreated, project)
}

// List handles GET /api/v1/projects.
func (h *ProjectsHandler) List(w http.ResponseWriter, r *http.Request) {
	projects, err := h.projects.List(r.Context())
	if err != nil {
		render.Error(w, r, http.StatusInternalServerError, "store_error", "failed to list projects", err)
		return
	}
	if projects == nil {
		projects = []*domain.Project{}
	}
	render.JSON(w, http.StatusOK, projects)
}

// CreateAPIKey handles POST /api/v1/projects/{id}/keys.
func (h *ProjectsHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		render.Error(w, r, http.StatusBadRequest, "invalid_id", "project id must be a valid UUID", nil)
		return
	}

	if _, err := h.projects.GetByID(r.Context(), projectID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			render.Error(w, r, http.StatusNotFound, "not_found", "project not found", nil)
			return
		}
		render.Error(w, r, http.StatusInternalServerError, "store_error", "failed to retrieve project", err)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Error(w, r, http.StatusBadRequest, "invalid_body", "request body must be valid JSON", err)
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		render.Error(w, r, http.StatusUnprocessableEntity, "missing_fields", "name is required", nil)
		return
	}

	rawKey, err := generateKey()
	if err != nil {
		render.Error(w, r, http.StatusInternalServerError, "internal_error", "failed to generate API key", err)
		return
	}

	key := &domain.APIKey{
		ID:        uuid.New(),
		ProjectID: projectID,
		Label:     strings.TrimSpace(req.Name),
		KeyHash:   hashKey(rawKey),
		CreatedAt: time.Now().UTC(),
	}

	if err := h.apiKeys.Create(r.Context(), key); err != nil {
		render.Error(w, r, http.StatusInternalServerError, "store_error", "failed to create API key", err)
		return
	}

	// Return the plaintext key exactly once — it is not stored.
	render.JSON(w, http.StatusCreated, map[string]any{
		"id":         key.ID,
		"key":        rawKey,
		"project_id": key.ProjectID,
		"label":      key.Label,
		"created_at": key.CreatedAt,
	})
}

// RevokeAPIKey handles DELETE /api/v1/projects/{id}/keys/{kid}.
func (h *ProjectsHandler) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	keyID, err := uuid.Parse(chi.URLParam(r, "kid"))
	if err != nil {
		render.Error(w, r, http.StatusBadRequest, "invalid_id", "key id must be a valid UUID", nil)
		return
	}

	if err := h.apiKeys.Revoke(r.Context(), keyID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			render.Error(w, r, http.StatusNotFound, "not_found", "API key not found", nil)
			return
		}
		render.Error(w, r, http.StatusInternalServerError, "store_error", "failed to revoke API key", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func generateKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate key: %w", err)
	}
	return "custos_" + hex.EncodeToString(b), nil
}

func hashKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonAlphanumeric.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}
