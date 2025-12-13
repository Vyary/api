package server

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/andybalholm/brotli"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func Cors(next http.Handler) http.Handler {
	allowedOrigins := map[string]bool{
		"https://exile-profit.com":      true,
		"https://poe2.exile-profit.com": true,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func CompressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enc := strings.ToLower(r.Header.Get("Accept-Encoding"))
		var writer io.WriteCloser

		switch {
		case strings.Contains(enc, "br"):
			w.Header().Set("Content-Encoding", "br")
			writer = brotli.NewWriterLevel(w, brotli.BestSpeed)
		case strings.Contains(enc, "gzip"):
			w.Header().Set("Content-Encoding", "gzip")
			var err error
			writer, err = gzip.NewWriterLevel(w, gzip.BestSpeed)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
		}

		if writer != nil {
			w.Header().Add("Vary", "Accept-Encoding")
			defer writer.Close()
			w = &compressResponseWriter{ResponseWriter: w, Writer: writer}
		}

		next.ServeHTTP(w, r)
	})
}

type compressResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w *compressResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func TraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		ctx, span := tracer.Start(ctx, fmt.Sprintf("%s %s", r.Method, r.URL.Path), trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.path", r.URL.Path),
		)

		next.ServeHTTP(w, r.WithContext(ctx))

		span.SetStatus(codes.Ok, "")
	})
}
