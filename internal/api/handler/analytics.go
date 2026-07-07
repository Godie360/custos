package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/iPFSoftwares/custos/internal/domain"
	"github.com/iPFSoftwares/custos/internal/store"
)

// AnalyticsStore is the minimal interface needed by AnalyticsHandler.
type AnalyticsStore interface {
	List(ctx context.Context, filter store.ListIssuesFilter) ([]*domain.Issue, error)
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
	TotalIssues          int            `json:"total_issues"`
	ErrorCountBySeverity map[string]int `json:"error_count_by_severity"`
	TopServices          []serviceCount `json:"top_services"`
}

type serviceCount struct {
	Service         string `json:"service"`
	OccurrenceCount int    `json:"occurrence_count"`
}

// Summary handles GET /api/v1/analytics/summary.
func (h *AnalyticsHandler) Summary(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	filter := store.ListIssuesFilter{Limit: 500}
	if pid := q.Get("project_id"); pid != "" {
		if id, err := uuid.Parse(pid); err == nil {
			filter.ProjectID = id
		}
	}

	issues, err := h.issues.List(r.Context(), filter)
	if err != nil {
		http.Error(w, "failed to load issues", http.StatusInternalServerError)
		return
	}

	bySeverity := make(map[string]int)
	serviceOccurrences := make(map[string]int)

	for _, iss := range issues {
		bySeverity[iss.Severity]++
		serviceOccurrences[iss.Service] += iss.OccurrenceCount
	}

	// Top 5 services by occurrence count.
	type svc struct {
		name  string
		count int
	}
	svcs := make([]svc, 0, len(serviceOccurrences))
	for name, cnt := range serviceOccurrences {
		svcs = append(svcs, svc{name, cnt})
	}
	// Simple insertion sort — small N.
	for i := 1; i < len(svcs); i++ {
		for j := i; j > 0 && svcs[j].count > svcs[j-1].count; j-- {
			svcs[j], svcs[j-1] = svcs[j-1], svcs[j]
		}
	}
	topN := 5
	if len(svcs) < topN {
		topN = len(svcs)
	}
	top := make([]serviceCount, topN)
	for i := 0; i < topN; i++ {
		top[i] = serviceCount{Service: svcs[i].name, OccurrenceCount: svcs[i].count}
	}

	resp := summaryResponse{
		TotalIssues:          len(issues),
		ErrorCountBySeverity: bySeverity,
		TopServices:          top,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
