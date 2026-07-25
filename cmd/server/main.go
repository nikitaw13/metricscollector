package main

import (
	"log"
	"net/http"
	"time"

	"github.com/PrometheRus/metricscollector/internal/handler"
	"github.com/PrometheRus/metricscollector/internal/repository"
	"github.com/go-chi/chi"
)

func main() {
	var s = repository.NewMemStorage()
	var h = handler.MetricsHandler{Storage: s}

	router := chi.NewRouter()

	router.Get("/", h.GetRootHandler)
	router.Get("/value/{TYPE}/{METRIC}", h.GetMetricHandler)

	router.Post("/update", h.PostNoTypeHandler)
	router.Route("/update/{TYPE}", func(r chi.Router) {
		r.Use(handler.TypeMiddleware)             // проверит TYPE для всех трёх ниже
		r.Post("/", h.PostNoMetricHandler)        // 404 — нет имени
		r.Post("/{METRIC}", h.PostNoValueHandler) // 400 — нет значения
		r.Post("/{METRIC}/{VALUE}", h.PostFullHandler)
	})

	srv := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      router,
	}

	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
