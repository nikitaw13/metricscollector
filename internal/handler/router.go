package handler

import (
	"github.com/go-chi/chi"
)

func (h *MetricsHandler) NewRouter() *chi.Mux {
	router := chi.NewRouter()

	router.Get("/", h.GetRootHandler)
	router.Get("/value/{TYPE}/{METRIC}", h.GetMetricHandler)

	router.Post("/update", h.PostNoTypeHandler)
	router.Route("/update/{TYPE}", func(r chi.Router) {
		r.Use(TypeMiddleware)                     // проверит TYPE для всех трёх ниже
		r.Post("/", h.PostNoMetricHandler)        // 404 — нет имени
		r.Post("/{METRIC}", h.PostNoValueHandler) // 400 — нет значения
		r.Post("/{METRIC}/{VALUE}", h.PostFullHandler)
	})

	return router
}
