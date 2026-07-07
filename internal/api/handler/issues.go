package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/iPFSoftwares/custos/internal/api/render"
	"github.com/iPFSoftwares/custos/internal/domain"
	"github.com/iPFSoftwares/custos/internal/store"
)

// IssueStoreReader is the minimal read interface needed by IssuesHandler.
type IssueStoreReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Issue, error)
	Update(ctx context.Context, issue *domain.Issue) error
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
		Limit:       50,
	}

	if pid := q.Get("project_id"); pid != "" {
		if id, err := uuid.Parse(pid); err == nil {
			filter.ProjectID = id
		}
	}
	if l := q.Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 200 {
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
		render.Error(w, r, http.StatusInternalServerError, "store_error", "failed to list issues", err)
		return
	}

	if issues == nil {
		issues = []*domain.Issue{}
	}
	render.JSON(w, http.StatusOK, issues)
}

// GetByID handles GET /api/v1/issues/{id}.
func (h *IssuesHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		render.Error(w, r, http.StatusBadRequest, "invalid_id", "issue id must be a valid UUID", nil)
		return
	}

	issue, err := h.store.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			render.Error(w, r, http.StatusNotFound, "not_found", "issue not found", nil)
			return
		}
		render.Error(w, r, http.StatusInternalServerError, "store_error", "failed to retrieve issue", err)
		return
	}

	render.JSON(w, http.StatusOK, issue)
}

// Patch handles PATCH /api/v1/issues/{id} — updates mutable fields (status, severity).
func (h *IssuesHandler) Patch(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		render.Error(w, r, http.StatusBadRequest, "invalid_id", "issue id must be a valid UUID", nil)
		return
	}

	var body struct {
		Status   *string `json:"status"`
		Severity *string `json:"severity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		render.Error(w, r, http.StatusBadRequest, "invalid_body", "request body must be valid JSON", err)
		return
	}

	issue, err := h.store.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			render.Error(w, r, http.StatusNotFound, "not_found", "issue not found", nil)
			return
		}
		render.Error(w, r, http.StatusInternalServerError, "store_error", "failed to retrieve issue", err)
		return
	}

	if body.Status != nil {
		switch domain.IssueStatus(*body.Status) {
		case domain.IssueStatusOpen, domain.IssueStatusResolved, domain.IssueStatusAnalysisFailed:
			issue.Status = domain.IssueStatus(*body.Status)
		default:
			render.Error(w, r, http.StatusUnprocessableEntity, "invalid_status",
				"status must be one of: open, resolved, analysis_failed", nil)
			return
		}
	}
	if body.Severity != nil {
		switch *body.Severity {
		case "critical", "error", "warning", "info":
			issue.Severity = *body.Severity
		default:
			render.Error(w, r, http.StatusUnprocessableEntity, "invalid_severity",
				"severity must be one of: critical, error, warning, info", nil)
			return
		}
	}

	if err := h.store.Update(r.Context(), issue); err != nil {
		render.Error(w, r, http.StatusInternalServerError, "store_error", "failed to update issue", err)
		return
	}

	render.JSON(w, http.StatusOK, issue)
}
