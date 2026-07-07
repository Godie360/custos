package handler

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/iPFSoftwares/custos/internal/api/render"
	"github.com/iPFSoftwares/custos/internal/store"
)

// AnalyticsStore is the minimal interface needed by AnalyticsHandler.
type AnalyticsStore interface {
	TotalCount(ctx context.Context) (int, error)
	CountBySeverity(ctx context.Context) ([]store.SeverityCount, error)
	TopServices(ctx context.Context) ([]store.ServiceCount, error)
}

// AnalyticsHandler handles GET /api/v1/analytics/summary.
type AnalyticsHandler struct {
	issues AnalyticsStore
}

// NewAnalyticsHandler creates an AnalyticsHandler.
func NewAnalyticsHandler(issues AnalyticsStore) *AnalyticsHandler {
	return &AnalyticsHandler{issues: issues}
}

type summaryResponse struct {
	TotalIssues          int              `json:"total_issues"`
	ErrorCountBySeverity map[string]int   `json:"error_count_by_severity"`
	TopServices          []topServiceItem `json:"top_services"`
}

type topServiceItem struct {
	Service         string `json:"service"`
	OccurrenceCount int    `json:"occurrence_count"`
}

// Summary handles GET /api/v1/analytics/summary.
// All aggregation runs in the database — no in-memory slices.
func (h *AnalyticsHandler) Summary(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// project_id filter is parsed but forwarded to store methods that support it;
	// the current aggregation queries are global — expand them when needed.
	if pid := q.Get("project_id"); pid != "" {
		if _, err := uuid.Parse(pid); err != nil {
			render.Error(w, r, http.StatusBadRequest, "invalid_id", "project_id must be a valid UUID", nil)
			return
		}
	}

	total, err := h.issues.TotalCount(r.Context())
	if err != nil {
		render.Error(w, r, http.StatusInternalServerError, "store_error", "failed to load analytics", err)
		return
	}

	severityCounts, err := h.issues.CountBySeverity(r.Context())
	if err != nil {
		render.Error(w, r, http.StatusInternalServerError, "store_error", "failed to load analytics", err)
		return
	}

	topSvcs, err := h.issues.TopServices(r.Context())
	if err != nil {
		render.Error(w, r, http.StatusInternalServerError, "store_error", "failed to load analytics", err)
		return
	}

	bySeverity := make(map[string]int, len(severityCounts))
	for _, sc := range severityCounts {
		bySeverity[sc.Severity] = sc.Count
	}

	top := make([]topServiceItem, len(topSvcs))
	for i, s := range topSvcs {
		top[i] = topServiceItem{Service: s.Service, OccurrenceCount: s.Total}
	}

	render.JSON(w, http.StatusOK, summaryResponse{
		TotalIssues:          total,
		ErrorCountBySeverity: bySeverity,
		TopServices:          top,
	})
}
