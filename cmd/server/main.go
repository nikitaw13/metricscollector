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

	_ "net/http/pprof"

	"github.com/nikitaw13/metricscollector/internal/audit"
	"github.com/nikitaw13/metricscollector/internal/handler"
	"github.com/nikitaw13/metricscollector/internal/logger"
	"github.com/nikitaw13/metricscollector/internal/repository"
	"github.com/nikitaw13/metricscollector/internal/retry"
	"github.com/nikitaw13/metricscollector/internal/service"
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

	zapLogger, err := logger.NewLogger(cfg.logLevel)
	if err != nil {
		return err
	}

	var storage service.Repository
	var dbPinger service.DBPinger
	var dbRetryDelays = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	isSyncWrite := cfg.storeInterval == 0

	switch {
	// Database storage
	case cfg.databaseDSN != "":
		postgresStorage, err := repository.NewPostgresStorageFromDSN(cfg.databaseDSN, cfg.migrationPath, dbRetryDelays, zapLogger)

		if err != nil {
			zapLogger.Error("failed to initialize postgres storage", zap.Error(err))
			return err
		}
		defer postgresStorage.Close()

		storage = postgresStorage
		dbPinger = postgresStorage

	// Persistent storage
	case cfg.fileStoragePath != "":
		memStorage := repository.NewMemStorage()
		persistentStorage := repository.NewPersistentMemStorage(memStorage, cfg.fileStoragePath, isSyncWrite)

		if cfg.restore {
			err := persistentStorage.Restore()

			if errors.Is(err, os.ErrNotExist) {
				zapLogger.Warn("file does not exist", zap.Error(err))
			} else if err != nil {
				zapLogger.Error("failed to restore metrics from file", zap.Error(err))
			}
		}

		storage = persistentStorage
		dbPinger = persistentStorage

	// In-memory storage
	default:
		memStorage := repository.NewMemStorage()
		storage = memStorage
		dbPinger = memStorage
	}

	var (
		retryDelays = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
		httpClient  = &http.Client{Timeout: 5 * time.Second}
		retryClient = retry.New(retryDelays, httpClient, zapLogger)
		svc         = service.New(storage, dbPinger, zapLogger)
	)

	auditor, err := audit.New(cfg.auditFile, cfg.auditURL, retryClient)
	if err != nil {
		zapLogger.Error("failed to initialize auditor", zap.Error(err))
		return err
	}

	metricsHandler := handler.NewMetricsHandler(cfg.hashKey, svc, zapLogger, auditor)

	var (
		router = metricsHandler.Router()

		srv = &http.Server{
			Addr:         cfg.serverHTTPAddr,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
			Handler:      router,
		}
	)

	if persistentStorage, ok := storage.(*repository.PersistentMemStorage); ok {
		if !isSyncWrite {
			go persistentStorage.PeriodicSave(ctx, time.Duration(cfg.storeInterval)*time.Second)
		}
	}

	serverErr := make(chan error, 1)
	go func() {
		zapLogger.Info("server is running", zap.String("address", cfg.serverHTTPAddr), zap.String("log_level", cfg.logLevel))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	if cfg.pprofHTTPAddr != "" {
		go func() {
			if err := http.ListenAndServe(cfg.pprofHTTPAddr, nil); err != nil && !errors.Is(err, http.ErrServerClosed) {
				zapLogger.Warn("failed to start pprof server", zap.Error(err))
			}
		}()
	}

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
	}

	defer func() {
		if persistentStorage, ok := storage.(*repository.PersistentMemStorage); ok {
			if err := persistentStorage.Save(); err != nil {
				zapLogger.Error("error saving metrics on shutdown", zap.Error(err))
			}
		}
	}()

	defer func() {
		if err := auditor.Close(); err != nil {
			zapLogger.Error("error closing auditor on shutdown", zap.Error(err))
		}
	}()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
