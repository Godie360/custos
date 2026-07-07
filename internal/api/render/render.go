package render

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type errorBody struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	TraceID string `json:"trace_id,omitempty"`
}

// JSON writes v as JSON with the given status code.
// Encoding errors are swallowed — they are a server bug, not a client error.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("render: json encode", slog.String("error", err.Error()))
	}
}

// Error writes a JSON error response. The caller provides a machine-readable
// code (e.g. "not_found") and a human-readable message. The raw internal err
// is logged server-side and never sent to the client.
func Error(w http.ResponseWriter, r *http.Request, status int, code, message string, internal error) {
	if internal != nil {
		slog.ErrorContext(r.Context(), "request error",
			slog.String("code", code),
			slog.String("error", internal.Error()),
		)
	}
	JSON(w, status, errorBody{
		Error: message,
		Code:  code,
	})
}
