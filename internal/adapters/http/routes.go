package http

import (
	"expvar"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimdw "github.com/go-chi/chi/v5/middleware"
)

const BasePath = "/api/v1"

type Dependencies struct {
	TagCreator      TagCreator
	TagRetriever    TagRetriever
	MediaCreator    MediaCreator
	MediaFinalizer  MediaFinalizer
	MediaRetriever  MediaRetriever
	Logger          *slog.Logger
	MetricForwarder MetricsForwarder
}

func NewRouter(deps Dependencies) chi.Router {
	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {})
	r.Get("/debug/vars", expvar.Handler().ServeHTTP)

	apiRouter := chi.NewRouter()

	apiRouter.Use(
		metricsMiddleware(deps),
		chimdw.RequestID,
		chimdw.Recoverer,
		loggerMiddleware(deps),
		loggerRequestIDMiddleware(),
	)

	apiRouter.Post("/tags", HandlePostTags(deps.TagCreator))
	apiRouter.Get("/tags", HandleGetTags(deps.TagRetriever))
	apiRouter.Post("/media", HandlePostMedia(deps.MediaCreator))
	apiRouter.Get("/media/{id}", HandleGetMedia(deps.MediaRetriever))
	apiRouter.Post("/media/{id}/finalize", HandlePostFinalizeMedia(deps.MediaFinalizer))

	r.Mount(BasePath, apiRouter)
	return r
}
