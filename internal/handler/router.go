package handler

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

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
		r.Use(TypeMiddleware)              // проверит TYPE для всех трёх ниже
		r.Post("/", h.PostNoMetric)        // 404 — нет имени
		r.Post("/{METRIC}", h.PostNoValue) // 400 — нет значения
		r.Post("/{METRIC}/{VALUE}", h.PostFull)
	})

	return r
}
