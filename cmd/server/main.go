package main

import (
	"log"
	"net/http"
	"time"

	"github.com/PrometheRus/metricscollector/internal/handler"
	"github.com/PrometheRus/metricscollector/internal/repository"
)

func main() {
	parseFlags()

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var r = repository.NewMemStorage()
	var h = handler.MetricsHandler{Storage: r}
	router := h.New()

	srv := &http.Server{
		Addr:         flagRunAddr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      router,
	}

	log.Println("Running server on", flagRunAddr)
	return srv.ListenAndServe()
}
