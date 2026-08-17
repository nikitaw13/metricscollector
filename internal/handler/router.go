package handler

import (
	"github.com/go-chi/chi"
)

// New builds and returns a chi router with all metric API routes.
func (h *MetricsHandler) New() *chi.Mux {
	router := chi.NewRouter()
	router.Use(loggerMiddleware)

	router.Get("/", h.handleListMetrics)
	router.Route("/value/{TYPE}", func(r chi.Router) {
		r.Use(typeMiddleware)
		r.Get("/{METRIC}", h.handleGetMetric)
	})

	router.Post("/update", h.handleMissingType)
	router.Route("/update/{TYPE}", func(r chi.Router) {
		r.Use(typeMiddleware)                     // validates TYPE for all routes below
		r.Post("/", h.handleMissingMetric)        // 404 — metric name missing
		r.Post("/{METRIC}", h.handleMissingValue) // 400 — metric value missing
		r.Post("/{METRIC}/{VALUE}", h.handleSetMetric)
	})

	return router
}
