package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/iPFSoftwares/custos/internal/domain"
	"github.com/iPFSoftwares/custos/internal/store"
)

// IssueStoreReader is the minimal read interface needed by IssuesHandler.
type IssueStoreReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Issue, error)
	List(ctx context.Context, filter store.ListIssuesFilter) ([]*domain.Issue, error)
}

// IssuesHandler handles issue-related HTTP endpoints.
type IssuesHandler struct {
	store IssueStoreReader
}

// NewIssuesHandler creates an IssuesHandler.
func NewIssuesHandler(s IssueStoreReader) *IssuesHandler {
	return &IssuesHandler{store: s}
}

// List handles GET /api/v1/issues with optional query filters.
func (h *IssuesHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	filter := store.ListIssuesFilter{
		Service:     q.Get("service"),
		Environment: q.Get("environment"),
		Severity:    q.Get("severity"),
	}

	if pid := q.Get("project_id"); pid != "" {
		id, err := uuid.Parse(pid)
		if err == nil {
			filter.ProjectID = id
		}
	}

	if l := q.Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			filter.Limit = n
		}
	}
	if o := q.Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			filter.Offset = n
		}
	}

	issues, err := h.store.List(r.Context(), filter)
	if err != nil {
		http.Error(w, "failed to list issues", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(issues)
}

// GetByID handles GET /api/v1/issues/{id}.
func (h *IssuesHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	rawID := chi.URLParam(r, "id")
	id, err := uuid.Parse(rawID)
	if err != nil {
		http.Error(w, "invalid issue id", http.StatusBadRequest)
		return
	}

	issue, err := h.store.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "issue not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to retrieve issue", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(issue)
}
