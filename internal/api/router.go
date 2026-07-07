package api

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/iPFSoftwares/custos/internal/api/handler"
	"github.com/iPFSoftwares/custos/internal/api/middleware"
	"github.com/iPFSoftwares/custos/internal/store"
)

// RouterDeps bundles all handler and store dependencies needed to build the router.
type RouterDeps struct {
	APIKeys  store.APIKeyStore
	Projects store.ProjectStore
	Ingest   *handler.IngestHandler
	Issues   *handler.IssuesHandler
	Analytics *handler.AnalyticsHandler
	ProjectsH *handler.ProjectsHandler
}

// NewRouter constructs and returns a fully configured Chi router.
func NewRouter(deps RouterDeps) http.Handler {
	r := chi.NewRouter()

	// Global middleware.
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-Custos-Key"},
		MaxAge:         300,
	}))
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	// Health check — no auth required.
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// OpenAPI spec — served for Swagger UI.
	r.Get("/api/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		specPath := os.Getenv("OPENAPI_SPEC_PATH")
		if specPath == "" {
			specPath = "api/openapi.yaml"
		}
		http.ServeFile(w, r, specPath)
	})

	r.Route("/api/v1", func(r chi.Router) {
		// Ingest — requires valid API key.
		r.With(middleware.APIKey(deps.APIKeys, deps.Projects)).
			Post("/ingest", deps.Ingest.ServeHTTP)

		// Issues.
		r.Get("/issues", deps.Issues.List)
		r.Get("/issues/{id}", deps.Issues.GetByID)

		// Analytics.
		r.Get("/analytics/summary", deps.Analytics.Summary)

		// Projects.
		r.Post("/projects", deps.ProjectsH.Create)
		r.Get("/projects", deps.ProjectsH.List)
		r.Post("/projects/{id}/keys", deps.ProjectsH.CreateAPIKey)
		r.Delete("/projects/{id}/keys/{kid}", deps.ProjectsH.RevokeAPIKey)
	})

	return r
}
