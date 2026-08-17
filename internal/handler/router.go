package handler

import (
	"github.com/go-chi/chi"
)

func (h *MetricsHandler) New() *chi.Mux {
	router := chi.NewRouter()

	router.Get("/", h.HandleIndex)
	router.Route("/value/{TYPE}", func(r chi.Router) {
		r.Use(TypeMiddleware)
		r.Get("/{METRIC}", h.HandleGetMetric)
	})

	router.Post("/update", h.HandleMissingType)
	router.Route("/update/{TYPE}", func(r chi.Router) {
		r.Use(TypeMiddleware)                // проверит TYPE для всех трёх ниже
		r.Post("/", h.HandleMissingMetric)   // 404 — нет имени
		r.Post("/{METRIC}", h.HandleMissingValue) // 400 — нет значения
		r.Post("/{METRIC}/{VALUE}", h.HandleUpdate)
	})

	return router
}
