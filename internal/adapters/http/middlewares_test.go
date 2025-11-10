package http

//go:generate mockgen -destination=mocks/mock_metrics_forwarder.go -package=mocks github.com/peano88/medias/internal/adapters/http MetricsForwarder

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/peano88/medias/internal/adapters/http/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestMetricsMiddleware(t *testing.T) {
	tests := []struct {
		name            string
		setupRouter     func(chi.Router, http.Handler)
		path            string
		method          string
		handlerStatus   int
		handlerDelay    time.Duration
		setupMock       func(*mocks.MockMetricsForwarder)
		validateMetrics func(*testing.T, *mocks.MockMetricsForwarder)
	}{
		{
			name: "captures GET request with 200 status",
			setupRouter: func(r chi.Router, h http.Handler) {
				r.Get("/api/v1/tags", h.ServeHTTP)
			},
			path:          "/api/v1/tags",
			method:        http.MethodGet,
			handlerStatus: http.StatusOK,
			handlerDelay:  10 * time.Millisecond,
			setupMock: func(mf *mocks.MockMetricsForwarder) {
				mf.EXPECT().
					AddRequestHit("GET /api/v1/tags", http.StatusOK, gomock.Any()).
					DoAndReturn(func(pattern string, code int, duration time.Duration) error {
						// Verify duration is reasonable (at least 10ms due to handler delay)
						assert.GreaterOrEqual(t, duration.Milliseconds(), int64(10))
						return nil
					})
			},
		},
		{
			name: "captures POST request with 201 status",
			setupRouter: func(r chi.Router, h http.Handler) {
				r.Post("/api/v1/tags", h.ServeHTTP)
			},
			path:          "/api/v1/tags",
			method:        http.MethodPost,
			handlerStatus: http.StatusCreated,
			handlerDelay:  5 * time.Millisecond,
			setupMock: func(mf *mocks.MockMetricsForwarder) {
				mf.EXPECT().
					AddRequestHit("POST /api/v1/tags", http.StatusCreated, gomock.Any()).
					Return(nil)
			},
		},
		{
			name: "captures 404 status",
			setupRouter: func(r chi.Router, h http.Handler) {
				r.Get("/api/v1/tags", h.ServeHTTP)
			},
			path:          "/api/v1/tags",
			method:        http.MethodGet,
			handlerStatus: http.StatusNotFound,
			setupMock: func(mf *mocks.MockMetricsForwarder) {
				mf.EXPECT().
					AddRequestHit("GET /api/v1/tags", http.StatusNotFound, gomock.Any()).
					Return(nil)
			},
		},
		{
			name: "captures 500 status",
			setupRouter: func(r chi.Router, h http.Handler) {
				r.Get("/api/v1/tags", h.ServeHTTP)
			},
			path:          "/api/v1/tags",
			method:        http.MethodGet,
			handlerStatus: http.StatusInternalServerError,
			setupMock: func(mf *mocks.MockMetricsForwarder) {
				mf.EXPECT().
					AddRequestHit("GET /api/v1/tags", http.StatusInternalServerError, gomock.Any()).
					Return(nil)
			},
		},
		{
			name: "handles metric forwarder error gracefully",
			setupRouter: func(r chi.Router, h http.Handler) {
				r.Get("/api/v1/tags", h.ServeHTTP)
			},
			path:          "/api/v1/tags",
			method:        http.MethodGet,
			handlerStatus: http.StatusOK,
			setupMock: func(mf *mocks.MockMetricsForwarder) {
				mf.EXPECT().
					AddRequestHit("GET /api/v1/tags", http.StatusOK, gomock.Any()).
					Return(errors.New("metrics service unavailable"))
			},
		},
		{
			name: "captures route with path parameters",
			setupRouter: func(r chi.Router, h http.Handler) {
				r.Get("/api/v1/tags/{id}", h.ServeHTTP)
			},
			path:          "/api/v1/tags/123",
			method:        http.MethodGet,
			handlerStatus: http.StatusOK,
			setupMock: func(mf *mocks.MockMetricsForwarder) {
				// Should use route pattern, not actual path
				mf.EXPECT().
					AddRequestHit("GET /api/v1/tags/{id}", http.StatusOK, gomock.Any()).
					Return(nil)
			},
		},
		{
			name: "defaults to 200 when no status written",
			setupRouter: func(r chi.Router, h http.Handler) {
				r.Get("/api/v1/health", h.ServeHTTP)
			},
			path:          "/api/v1/health",
			method:        http.MethodGet,
			handlerStatus: 0, // Handler doesn't explicitly set status
			setupMock: func(mf *mocks.MockMetricsForwarder) {
				mf.EXPECT().
					AddRequestHit("GET /api/v1/health", http.StatusOK, gomock.Any()).
					Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMetrics := mocks.NewMockMetricsForwarder(ctrl)
			tt.setupMock(mockMetrics)

			logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
			deps := Dependencies{
				Logger:          logger,
				MetricForwarder: mockMetrics,
			}

			// Create test handler
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.handlerDelay > 0 {
					time.Sleep(tt.handlerDelay)
				}
				if tt.handlerStatus > 0 {
					w.WriteHeader(tt.handlerStatus)
				}
				_, _ = w.Write([]byte("ok"))
			})

			// Setup router with middleware
			r := chi.NewRouter()
			r.Use(metricsMiddleware(deps))
			tt.setupRouter(r, testHandler)

			// Make request
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			// Expectations are verified by gomock
		})
	}
}

func TestWrappedResponseWriter(t *testing.T) {
	t.Run("captures WriteHeader status", func(t *testing.T) {
		rec := httptest.NewRecorder()
		wrapped := newWrappedResponseWriter(rec)

		wrapped.WriteHeader(http.StatusCreated)

		assert.Equal(t, http.StatusCreated, wrapped.statusCode)
	})

	t.Run("defaults to 200 when Write called without WriteHeader", func(t *testing.T) {
		rec := httptest.NewRecorder()
		wrapped := newWrappedResponseWriter(rec)

		written, err := wrapped.Write([]byte("test"))

		assert.Equal(t, http.StatusOK, wrapped.statusCode)
		assert.Equal(t, 4, written)
		assert.NoError(t, err)
	})

	t.Run("preserves status when Write called after WriteHeader", func(t *testing.T) {
		rec := httptest.NewRecorder()
		wrapped := newWrappedResponseWriter(rec)

		wrapped.WriteHeader(http.StatusAccepted)
		written, err := wrapped.Write([]byte("test"))

		assert.Equal(t, http.StatusAccepted, wrapped.statusCode)
		assert.Equal(t, 4, written)
		assert.NoError(t, err)
	})
}
