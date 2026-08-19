package main

import (
	"errors"
	"log"
	"net/http"
	"os"
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
	if err := handler.InitLogger(flagLogLevel); err != nil {
		return err
	}

	isSynchronic := flagStoreInterval == 0
	memStorage := repository.NewMemStorage()
	persistentMemStorage := repository.NewPersistentMemStorage(memStorage, flagFileStoragePath, isSynchronic)

	if flagRestore {
		err := persistentMemStorage.Restore()

		if errors.Is(err, os.ErrNotExist) {
			handler.Logger.Warn("File not exist", zap.Error(err))
		} else if err != nil {
			handler.Logger.Error("Error while restoring the file", zap.Error(err))
		}
	}

	metricsHandler := handler.MetricsHandler{Storage: persistentMemStorage}
	router := metricsHandler.New()

	srv := &http.Server{
		Addr:         flagHTTPAddr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      router,
	}

	if !isSynchronic {
		go persistentMemStorage.PeriodicSave(flagStoreInterval)
	}

	handler.Logger.Info("Running server", zap.String("address", flagHTTPAddr), zap.String("logLevel", flagLogLevel))
	return srv.ListenAndServe()
}
