package handler

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// Router builds and returns a chi router with all metric API routes.
func (h *MetricsHandler) Router() *chi.Mux {
	router := chi.NewRouter()
	router.Use(loggingMiddleware(h.logger))
	router.Use(decompressMiddleware)
	if h.hashKey != "" {
		router.Use(h.validateHashMiddleware)
	}
	router.Use(compressMiddleware)
	if h.hashKey != "" {
		router.Use(h.writeHashHeaderMiddleware)
	}

	router.Get("/", h.handleListMetrics)
	router.Route("/value/{type}", func(r chi.Router) {
		r.Use(requireMetricType)
		r.Get("/{metric}", h.handleURLRead)
	})

	router.Route("/update", func(r chi.Router) {
		r.Use(requireJSONContent)
		r.Post("/", h.handleJSONUpdate)
	})

	router.Route("/updates", func(r chi.Router) {
		r.Use(requireJSONContent)
		r.Post("/", h.handleJSONUpdates)
	})

	router.Route("/value", func(r chi.Router) {
		r.Use(middleware.AllowContentType("application/json"))
		r.Post("/", h.handleJSONRead)
	})

	router.Route("/update/{type}", func(r chi.Router) {
		r.Use(requireMetricType)                  // validates type for all routes below
		r.Post("/", h.handleMissingMetric)        // 404 — metric name missing
		r.Post("/{metric}", h.handleMissingValue) // 400 — metric value missing
		r.Post("/{metric}/{value}", h.handleURLUpdate)
	})

	router.Get("/ping", h.handleDatabasePing)

	return router
}
