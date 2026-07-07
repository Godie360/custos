package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/iPFSoftwares/custos/internal/api/middleware"
	"github.com/iPFSoftwares/custos/internal/api/render"
	"github.com/iPFSoftwares/custos/internal/domain"
	"github.com/iPFSoftwares/custos/pkg/event"
)

// IngestionService is the minimal interface consumed by IngestHandler.
type IngestionService interface {
	Ingest(ctx context.Context, raw *domain.RawEvent) error
}

// IngestHandler handles POST /api/v1/ingest.
type IngestHandler struct {
	svc IngestionService
}

// NewIngestHandler creates an IngestHandler backed by the given service.
func NewIngestHandler(svc IngestionService) *IngestHandler {
	return &IngestHandler{svc: svc}
}

// ServeHTTP decodes the SDK payload, validates it, and calls the ingestion service.
func (h *IngestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var p event.Payload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		render.Error(w, r, http.StatusBadRequest, "invalid_body", "request body must be valid JSON", err)
		return
	}

	if p.Service == "" || p.ErrorType == "" || p.Message == "" {
		render.Error(w, r, http.StatusUnprocessableEntity, "missing_fields",
			"service, error_type, and message are required", nil)
		return
	}

	env := p.Environment
	if env == "" {
		env = "production"
	}

	project := middleware.GetProject(r.Context())
	if project == nil {
		render.Error(w, r, http.StatusInternalServerError, "internal_error", "project context missing", nil)
		return
	}

	raw := &domain.RawEvent{
		ID:          uuid.New(),
		ProjectID:   project.ID,
		Service:     p.Service,
		Environment: env,
		ErrorType:   p.ErrorType,
		Message:     p.Message,
		StackTrace:  p.StackTrace,
		ReceivedAt:  time.Now().UTC(),
	}

	if err := h.svc.Ingest(r.Context(), raw); err != nil {
		render.Error(w, r, http.StatusInternalServerError, "ingest_failed", "failed to ingest event", err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
