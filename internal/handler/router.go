package handler

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// New builds and returns the chi router with all registered routes.
func (h *MetricsHandler) New() *chi.Mux {
	r := chi.NewRouter()
	r.Use(LoggerMiddleware)

	r.Get("/", h.GetRoot)
	r.Route("/value/{TYPE}", func(r chi.Router) {
		r.Use(middleware.AllowContentType("text/plain"))
		r.Use(TypeMiddleware)
		r.Get("/{METRIC}", h.GetMetric)
	})

	r.Route("/update", func(r chi.Router) {
		r.Use(middleware.AllowContentType("application/json"))
		r.Post("/", h.PostUpdate)
	})

	r.Route("/value", func(r chi.Router) {
		r.Use(middleware.AllowContentType("application/json"))
		r.Post("/", h.PostValue)
	})

	r.Route("/update/{TYPE}", func(r chi.Router) {
		r.Use(middleware.AllowContentType("text/plain"))
		r.Use(TypeMiddleware)                           // validates TYPE for all three routes below
		r.Post("/", h.HandleUpdateMissingName)          // 404 — metric name missing
		r.Post("/{METRIC}", h.HandleUpdateMissingValue) // 400 — metric value missing
		r.Post("/{METRIC}/{VALUE}", h.HandleUpdateParam)
	})

	return r
}
