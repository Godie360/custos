package handler

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/iPFSoftwares/custos/internal/domain"
)

// systemOwnerID is the fixed seed user used as project owner before auth is implemented.
var systemOwnerID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonAlphanumeric.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	project := &domain.Project{
		ID:        uuid.New(),
		Name:      req.Name,
		Slug:      slugify(req.Name),
		OwnerID:   systemOwnerID,
		CreatedAt: time.Now().UTC(),
	}

	if err := h.projects.Create(r.Context(), project); err != nil {
		http.Error(w, "failed to create project", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(project)
}

// List handles GET /api/v1/projects.
func (h *ProjectsHandler) List(w http.ResponseWriter, r *http.Request) {
	projects, err := h.projects.List(r.Context())
	if err != nil {
		http.Error(w, "failed to list projects", http.StatusInternalServerError)
		return
	}
	if projects == nil {
		projects = []*domain.Project{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
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
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	rawKey, err := generateKey()
	if err != nil {
		http.Error(w, "failed to generate API key", http.StatusInternalServerError)
		return
	}

	key := &domain.APIKey{
		ID:        uuid.New(),
		ProjectID: projectID,
		Label:     req.Name,
		KeyHash:   hashKey(rawKey),
		CreatedAt: time.Now().UTC(),
	}

	if err := h.apiKeys.Create(r.Context(), key); err != nil {
		http.Error(w, "failed to create API key", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
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

func generateKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "custos_" + hex.EncodeToString(b), nil
}

func hashKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
