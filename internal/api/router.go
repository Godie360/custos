package api

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/iPFSoftwares/custos/internal/api/handler"
	"github.com/iPFSoftwares/custos/internal/api/middleware"
	"github.com/iPFSoftwares/custos/internal/api/render"
	"github.com/iPFSoftwares/custos/internal/store"
)

const maxBodyBytes = 1 << 20 // 1 MiB

// RouterDeps bundles all handler and store dependencies needed to build the router.
type RouterDeps struct {
	DB        *sql.DB
	APIKeys   store.APIKeyStore
	Projects  store.ProjectStore
	Ingest    *handler.IngestHandler
	Issues    *handler.IssuesHandler
	Analytics *handler.AnalyticsHandler
	ProjectsH *handler.ProjectsHandler
}

// NewRouter constructs and returns a fully configured Chi router.
func NewRouter(deps RouterDeps) http.Handler {
	r := chi.NewRouter()

	// Global middleware stack — order matters.
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-Custos-Key", "X-Request-ID"},
		ExposedHeaders: []string{"X-Request-ID"},
		MaxAge:         300,
	}))
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(chimiddleware.RequestSize(maxBodyBytes))

	// 404 / 405 with JSON bodies instead of plain text.
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		render.Error(w, r, http.StatusNotFound, "not_found", "the requested resource does not exist", nil)
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		render.Error(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
	})

	// Health check — pings the database so load balancers get a real signal.
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := deps.DB.PingContext(r.Context()); err != nil {
			render.Error(w, r, http.StatusServiceUnavailable, "db_unreachable", "database is not reachable", err)
			return
		}
		render.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
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
		r.Patch("/issues/{id}", deps.Issues.Patch)

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
