package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"

	"github.com/cgund98/go-postgres-api-template/internal/observability"
)

// RequestLogger returns a middleware that logs HTTP requests
func RequestLogger(rootLogger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			requestID := uuid.New()
			reqLogger := rootLogger.With("request_id", requestID)

			spanCtx := trace.SpanFromContext(r.Context()).SpanContext()
			if spanCtx.IsValid() {
				reqLogger = reqLogger.With(
					"trace_id", spanCtx.TraceID().String(),
					"span_id", spanCtx.SpanID().String(),
				)
			}

			ctx := r.Context()
			ctx = observability.SetLoggerOnContext(ctx, reqLogger)
			r = r.WithContext(ctx)

			// Wrap the response writer to capture status code
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Process the request
			next.ServeHTTP(ww, r)

			// Log the request
			duration := time.Since(start)
			// Use only the path without query parameters for security/privacy
			// r.URL.Path already excludes query parameters (which are in r.URL.RawQuery)
			reqLogger.InfoContext(ctx, "HTTP request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"duration_ms", duration.Milliseconds(),
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
		})
	}
}

func TracerSetter(tracer trace.Tracer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = observability.SetTracerOnContext(ctx, tracer)
			r = r.WithContext(ctx)

			// Wrap the response writer to capture status code
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Process the request
			next.ServeHTTP(ww, r)
		})
	}
}
