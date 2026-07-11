package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Godie360/custos/internal/api/render"
	"github.com/Godie360/custos/internal/domain"
)

// FilterRuleStore is the minimal interface needed by FiltersHandler.
type FilterRuleStore interface {
	Create(ctx context.Context, rule *domain.FilterRule) error
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]*domain.FilterRule, error)
	Delete(ctx context.Context, id uuid.UUID, projectID uuid.UUID) error
}

// FiltersHandler handles CRUD for project-scoped ingest filter rules.
type FiltersHandler struct {
	filters  FilterRuleStore
	projects ProjectsStore
}

// NewFiltersHandler creates a FiltersHandler.
func NewFiltersHandler(filters FilterRuleStore, projects ProjectsStore) *FiltersHandler {
	return &FiltersHandler{filters: filters, projects: projects}
}

type createFilterRequest struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

// Create handles POST /api/v1/projects/:id/filters
func (h *FiltersHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		render.Error(w, r, http.StatusBadRequest, "invalid_id", "invalid project id", nil)
		return
	}

	if _, err := h.projects.GetByID(r.Context(), projectID); err != nil {
		render.Error(w, r, http.StatusNotFound, "not_found", "project not found", nil)
		return
	}

	var req createFilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Error(w, r, http.StatusBadRequest, "invalid_body", "request body is not valid JSON", nil)
		return
	}

	if !validField(req.Field) {
		render.Error(w, r, http.StatusBadRequest, "invalid_field",
			"field must be one of: error_type, message, service, environment", nil)
		return
	}
	if !validOperator(req.Operator) {
		render.Error(w, r, http.StatusBadRequest, "invalid_operator",
			"operator must be one of: equals, contains, starts_with", nil)
		return
	}
	if req.Value == "" {
		render.Error(w, r, http.StatusBadRequest, "missing_value", "value is required", nil)
		return
	}

	rule := &domain.FilterRule{
		ProjectID: projectID,
		Field:     domain.FilterField(req.Field),
		Operator:  domain.FilterOperator(req.Operator),
		Value:     req.Value,
	}
	if err := h.filters.Create(r.Context(), rule); err != nil {
		render.Error(w, r, http.StatusInternalServerError, "internal", "failed to create filter rule", err)
		return
	}
	render.JSON(w, http.StatusCreated, rule)
}

// List handles GET /api/v1/projects/:id/filters
func (h *FiltersHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		render.Error(w, r, http.StatusBadRequest, "invalid_id", "invalid project id", nil)
		return
	}

	rules, err := h.filters.ListByProject(r.Context(), projectID)
	if err != nil {
		render.Error(w, r, http.StatusInternalServerError, "internal", "failed to list filter rules", err)
		return
	}
	if rules == nil {
		rules = []*domain.FilterRule{}
	}
	render.JSON(w, http.StatusOK, rules)
}

// Delete handles DELETE /api/v1/projects/:id/filters/:fid
func (h *FiltersHandler) Delete(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		render.Error(w, r, http.StatusBadRequest, "invalid_id", "invalid project id", nil)
		return
	}

	filterID, err := uuid.Parse(chi.URLParam(r, "fid"))
	if err != nil {
		render.Error(w, r, http.StatusBadRequest, "invalid_id", "invalid filter rule id", nil)
		return
	}

	if err := h.filters.Delete(r.Context(), filterID, projectID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			render.Error(w, r, http.StatusNotFound, "not_found", "filter rule not found", nil)
			return
		}
		render.Error(w, r, http.StatusInternalServerError, "internal", "failed to delete filter rule", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func validField(f string) bool {
	switch domain.FilterField(f) {
	case domain.FilterFieldErrorType, domain.FilterFieldMessage,
		domain.FilterFieldService, domain.FilterFieldEnvironment:
		return true
	}
	return false
}

func validOperator(o string) bool {
	switch domain.FilterOperator(o) {
	case domain.FilterOperatorEquals, domain.FilterOperatorContains,
		domain.FilterOperatorStartsWith:
		return true
	}
	return false
}
