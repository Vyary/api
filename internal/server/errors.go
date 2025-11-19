package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Errors map[string]string

type AppErr struct {
	Status   int    `json:"status"`
	Title    string `json:"title"`
	Details  string `json:"details"`
	Instance string `json:"instance"`
	Errors   Errors `json:"errors,omitempty"`
}

func CaptureError(ctx context.Context, message string, err error) {
	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID().String()

	slog.Error(message, "error", err, "traceID", traceID)

	span.SetStatus(codes.Error, message)
	span.RecordError(err)
}

func WriteJSON(ctx context.Context, w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		CaptureError(ctx, "encoding response", err)
	}
}

func NewInternalError(ctx context.Context, w http.ResponseWriter, message string, err error, path string) {
	appErr := AppErr{
		Status:   http.StatusInternalServerError,
		Title:    "Internal Server Error",
		Details:  "An unexpected internal server error occurred.",
		Instance: path,
	}
	WriteJSON(ctx, w, http.StatusInternalServerError, appErr)

	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID().String()

	slog.Error(message, "error", err, "traceID", traceID)

	span.SetStatus(codes.Error, message)
	span.RecordError(err)
}
