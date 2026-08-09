package handler

import (
	"github.com/go-chi/chi"
)

func (h *MetricsHandler) New() *chi.Mux {
	router := chi.NewRouter()
	router.Use(LoggerMiddleware)

	router.Get("/", h.GetRoot)
	router.Route("/value/{TYPE}", func(r chi.Router) {
		r.Use(TypeMiddleware)
		r.Get("/{METRIC}", h.GetMetric)
	})

	router.Post("/update", h.PostNoType)
	router.Route("/update/{TYPE}", func(r chi.Router) {
		r.Use(TypeMiddleware)              // проверит TYPE для всех трёх ниже
		r.Post("/", h.PostNoMetric)        // 404 — нет имени
		r.Post("/{METRIC}", h.PostNoValue) // 400 — нет значения
		r.Post("/{METRIC}/{VALUE}", h.PostFull)
	})

	return router
}
