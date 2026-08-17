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

	storage := repository.New()
	metricsHandler := handler.MetricsHandler{Storage: storage}
	router := metricsHandler.New()

	srv := &http.Server{
		Addr:         flagHTTPAddr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      router,
	}

	handler.Logger.Info("Running server", zap.String("address", flagHTTPAddr), zap.String("logLevel", flagLogLevel))
	return srv.ListenAndServe()
}
