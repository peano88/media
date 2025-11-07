package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimdw "github.com/go-chi/chi/v5/middleware"
)

const BasePath = "/api/v1"

type Dependencies struct {
	TagCreator      TagCreator
	Logger          *slog.Logger
	MetricForwarder MetricsForwarder
}

func NewRouter(deps Dependencies) chi.Router {
	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {})

	apiRouter := chi.NewRouter()

	apiRouter.Use(
		metricsMiddleware(deps),
		chimdw.RequestID,
		chimdw.Recoverer,
		loggerMiddleware(deps),
		loggerRequestIDMiddleware(),
	)

	apiRouter.Post("/tags", HandlePostTags(deps.TagCreator))

	r.Mount(BasePath, apiRouter)
	return r
}
