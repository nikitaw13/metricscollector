package main

import (
	"log"
	"net/http"
	"time"

	"github.com/PrometheRus/metricscollector/internal/handler"
	"github.com/PrometheRus/metricscollector/internal/repository"
	"go.uber.org/zap"
)

func main() {
	parseFlags()
	parseEnvs()

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if err := handler.Initialize(flagLogLevel); err != nil {
		return err
	}

	var r = repository.New()
	var h = handler.MetricsHandler{Storage: r}
	router := h.New()

	srv := &http.Server{
		Addr:         flagRunAddr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      router,
	}

	handler.Log.Info("Running server", zap.String("address", flagRunAddr), zap.String("logLevel", flagLogLevel))
	return srv.ListenAndServe()
}
