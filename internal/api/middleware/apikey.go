package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/iPFSoftwares/custos/internal/domain"
	"github.com/iPFSoftwares/custos/internal/store"
)

type projectContextKey string

const projectKey projectContextKey = "project"

// APIKey validates the X-Custos-Key header against the API key store.
// On success it attaches the project to the request context.
func APIKey(apiKeys store.APIKeyStore, projects store.ProjectStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := r.Header.Get("X-Custos-Key")
			if raw == "" {
				http.Error(w, "missing X-Custos-Key header", http.StatusUnauthorized)
				return
			}

			hash := hashKey(raw)
			key, err := apiKeys.GetByHash(r.Context(), hash)
			if err != nil {
				http.Error(w, "invalid API key", http.StatusUnauthorized)
				return
			}

			if key.RevokedAt != nil {
				http.Error(w, "API key has been revoked", http.StatusUnauthorized)
				return
			}

			if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
				http.Error(w, "API key has expired", http.StatusUnauthorized)
				return
			}

			project, err := projects.GetByID(r.Context(), key.ProjectID)
			if err != nil {
				http.Error(w, "project not found", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), projectKey, project)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetProject retrieves the project attached by the APIKey middleware.
func GetProject(ctx context.Context) *domain.Project {
	if p, ok := ctx.Value(projectKey).(*domain.Project); ok {
		return p
	}
	return nil
}

func hashKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
