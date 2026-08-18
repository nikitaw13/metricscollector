package handler

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// New builds and returns a chi router with all metric API routes.
func (h *MetricsHandler) New() *chi.Mux {
	router := chi.NewRouter()
	router.Use(loggerMiddleware)

	router.Get("/", h.handleListMetrics)
	router.Route("/value/{TYPE}", func(r chi.Router) {
		r.Use(requireMetricType)
		r.Get("/{METRIC}", h.handleURLRead)
	})

	router.Route("/update", func(r chi.Router) {
		r.Use(requireJSONContent)
		r.Post("/", h.handleJSONUpdate)
	})

	router.Route("/value", func(r chi.Router) {
		r.Use(middleware.AllowContentType("application/json"))
		r.Post("/", h.handleJSONRead)
	})

	router.Route("/update/{TYPE}", func(r chi.Router) {
		r.Use(requireMetricType)                  // validates TYPE for all routes below
		r.Post("/", h.handleMissingMetric)        // 404 — metric name missing
		r.Post("/{METRIC}", h.handleMissingValue) // 400 — metric value missing
		r.Post("/{METRIC}/{VALUE}", h.handleURLUpdate)
	})

	return router
}
