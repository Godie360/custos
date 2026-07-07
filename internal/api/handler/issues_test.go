package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Godie360/custos/internal/api/handler"
	"github.com/Godie360/custos/internal/domain"
	"github.com/Godie360/custos/internal/store"
)

// stubIssueStore is a thread-safe in-memory stub satisfying handler.IssueStoreReader.
type stubIssueStore struct {
	mu     sync.RWMutex
	issues map[uuid.UUID]*domain.Issue
}

func (s *stubIssueStore) GetByID(_ context.Context, id uuid.UUID) (*domain.Issue, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if issue, ok := s.issues[id]; ok {
		return issue, nil
	}
	return nil, domain.ErrNotFound
}

func (s *stubIssueStore) Update(_ context.Context, issue *domain.Issue) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.issues[issue.ID]; !ok {
		return domain.ErrNotFound
	}
	s.issues[issue.ID] = issue
	return nil
}

func (s *stubIssueStore) List(_ context.Context, _ store.ListIssuesFilter) ([]*domain.Issue, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*domain.Issue, 0, len(s.issues))
	for _, v := range s.issues {
		result = append(result, v)
	}
	return result, nil
}

func newStub(issues ...*domain.Issue) *stubIssueStore {
	m := make(map[uuid.UUID]*domain.Issue, len(issues))
	for _, iss := range issues {
		m[iss.ID] = iss
	}
	return &stubIssueStore{issues: m}
}

// chiContext injects a chi URL param into the request context so handlers
// that call chi.URLParam work correctly without a real router.
func chiContext(r *http.Request, key, val string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, val)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestIssuesHandler_GetByID(t *testing.T) {
	t.Parallel()

	knownID := uuid.New()

	tests := []struct {
		name       string
		id         string
		wantStatus int
	}{
		{"existing issue returns 200", knownID.String(), http.StatusOK},
		{"unknown id returns 404", uuid.New().String(), http.StatusNotFound},
		{"malformed id returns 400", "not-a-uuid", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Each subtest gets its own handler + store to avoid shared map races.
			h := handler.NewIssuesHandler(newStub(&domain.Issue{
				ID:      knownID,
				Service: "auth",
				Status:  domain.IssueStatusOpen,
			}))

			req := httptest.NewRequest(http.MethodGet, "/api/v1/issues/"+tt.id, nil)
			req = chiContext(req, "id", tt.id)
			rec := httptest.NewRecorder()

			h.GetByID(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestIssuesHandler_Patch(t *testing.T) {
	t.Parallel()

	knownID := uuid.New()

	tests := []struct {
		name       string
		id         string
		body       string
		wantStatus int
	}{
		{
			name:       "resolve existing issue",
			id:         knownID.String(),
			body:       `{"status":"resolved"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid status returns 422",
			id:         knownID.String(),
			body:       `{"status":"invalid"}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "unknown id returns 404",
			id:         uuid.New().String(),
			body:       `{"status":"resolved"}`,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "malformed id returns 400",
			id:         "bad",
			body:       `{"status":"resolved"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Each subtest gets its own store so parallel writes don't race.
			h := handler.NewIssuesHandler(newStub(&domain.Issue{
				ID:     knownID,
				Status: domain.IssueStatusOpen,
			}))

			req := httptest.NewRequest(http.MethodPatch, "/api/v1/issues/"+tt.id,
				strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req = chiContext(req, "id", tt.id)
			rec := httptest.NewRecorder()

			h.Patch(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code, rec.Body.String())
		})
	}
}

func TestIssuesHandler_List(t *testing.T) {
	t.Parallel()

	h := handler.NewIssuesHandler(newStub(
		&domain.Issue{ID: uuid.New(), Service: "api"},
		&domain.Issue{ID: uuid.New(), Service: "worker"},
	))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/issues", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	// List returns a JSON array directly.
	var issues []domain.Issue
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&issues))
	assert.Len(t, issues, 2)
}
