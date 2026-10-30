package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nikitaw13/metricscollector/internal/handler"
	"github.com/nikitaw13/metricscollector/internal/repository"
	"go.uber.org/zap"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	if err := run(cfg); err != nil {
		log.Fatal(err)
	}
}

func run(cfg Config) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := handler.InitLogger(cfg.logLevel); err != nil {
		return err
	}

	var storageToUse handler.Repository

	var dbToUse handler.DBPinger
	var timeouts = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	isSyncWrite := cfg.storeInterval == 0

	switch {
	// Database storage
	case cfg.databaseDSN != "":
		ps, err := repository.NewPostgresStorageFromDSN(cfg.databaseDSN, cfg.migrationPath, timeouts, handler.Logger)

		if err != nil {
			handler.Logger.Error("failed to initialize postgres storage", zap.Error(err))
			return err
		}
		defer ps.Close()

		storageToUse = ps
		dbToUse = ps

	// Persistent storage
	case cfg.fileStoragePath != "":
		memStorage := repository.NewMemStorage()
		persistentMemStorage := repository.NewPersistentMemStorage(memStorage, cfg.fileStoragePath, isSyncWrite)

		if cfg.restore {
			err := persistentMemStorage.Restore()

			if errors.Is(err, os.ErrNotExist) {
				handler.Logger.Warn("file does not exist", zap.Error(err))
			} else if err != nil {
				handler.Logger.Error("failed to restore metrics from file", zap.Error(err))
			}
		}

		storageToUse = persistentMemStorage
		dbToUse = persistentMemStorage

	// In-memory storage
	default:
		memStorage := repository.NewMemStorage()
		storageToUse = memStorage
		dbToUse = memStorage
	}

	metricsHandler := handler.NewMetricsHandler(storageToUse, dbToUse, cfg.hashKey)

	router := metricsHandler.NewRouter()

	srv := &http.Server{
		Addr:         cfg.serverHTTPAddr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      router,
	}

	if persistentStorage, ok := storageToUse.(*repository.PersistentMemStorage); ok {
		if !isSyncWrite {
			go persistentStorage.PeriodicSave(ctx, time.Duration(cfg.storeInterval)*time.Second)
		}
	}

	serverErr := make(chan error, 1)
	go func() {
		handler.Logger.Info("server is running", zap.String("address", cfg.serverHTTPAddr), zap.String("log_level", cfg.logLevel))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
	}

	defer func() {
		if persistentStorage, ok := storageToUse.(*repository.PersistentMemStorage); ok {
			if err := persistentStorage.Save(); err != nil {
				handler.Logger.Error("error saving metrics on shutdown", zap.Error(err))
			}
		}
	}()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
