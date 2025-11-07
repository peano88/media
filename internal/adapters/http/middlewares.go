package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimdw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"
)

type middlewarehandler func(http.Handler) http.Handler

func loggerMiddleware(deps Dependencies) middlewarehandler {
	return httplog.RequestLogger(deps.Logger, &httplog.Options{
		// Level defines the verbosity of the request logs:
		// slog.LevelDebug - log all responses (incl. OPTIONS)
		// slog.LevelInfo  - log all responses (excl. OPTIONS)
		// slog.LevelWarn  - log 4xx and 5xx responses only (except for 429)
		// slog.LevelError - log 5xx responses only
		Level: slog.LevelInfo,

		// Log attributes using given schema/format.

		// RecoverPanics recovers from panics occurring in the underlying HTTP handlers
		// and middlewares. It returns HTTP 500 unless response status was already set.
		//
		// NOTE: Panics are logged as errors automatically, regardless of this setting.
		RecoverPanics: true,
	})
}

func loggerRequestIDMiddleware() middlewarehandler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add request ID to the logger fields
			httplog.SetAttrs(r.Context(), slog.String("request_id", chimdw.GetReqID(r.Context())))

			next.ServeHTTP(w, r)
		})
	}
}

type MetricsForwarder interface {
	AddRequestHit(string, int, time.Duration) error
}

type wrappedResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (wr *wrappedResponseWriter) WriteHeader(code int) {
	wr.statusCode = code
	wr.ResponseWriter.WriteHeader(code)
}

func (wr *wrappedResponseWriter) Write(b []byte) (int, error) {
	if wr.statusCode == 0 {
		wr.statusCode = http.StatusOK
	}
	return wr.ResponseWriter.Write(b)
}

func newWrappedResponseWriter(w http.ResponseWriter) *wrappedResponseWriter {
	return &wrappedResponseWriter{ResponseWriter: w}
}

func metricsMiddleware(deps Dependencies) middlewarehandler {
	// metric is the first middleware to be executed,
	// and we should not rely on any other middleware to be executed before it.
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			before := time.Now()
			wrw := newWrappedResponseWriter(w)
			next.ServeHTTP(wrw, r)
			after := time.Now()

			chiCtx, ok := r.Context().Value(chi.RouteCtxKey).(*chi.Context)
			if !ok {
				deps.Logger.Error("failed to get chi route context",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path))
				return

			}
			route := r.Method + " " + chiCtx.RoutePattern()
			if err := deps.MetricForwarder.AddRequestHit(route, wrw.statusCode, after.Sub(before)); err != nil {
				deps.Logger.Error("failed to forward metrics", slog.String("pattern", route),
					slog.Int("code", wrw.statusCode),
					slog.Duration("duration", after.Sub(before)),
					slog.String("error", err.Error()))
			}

		})
	}
}
