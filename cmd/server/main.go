package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

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

	isSyncWrite := flagStoreInterval == 0
	memStorage := repository.NewMemStorage()
	persistentMemStorage := repository.NewPersistentMemStorage(memStorage, flagFileStoragePath, isSyncWrite)

	if flagRestore {
		err := persistentMemStorage.Restore()

		if errors.Is(err, os.ErrNotExist) {
			handler.Logger.Warn("File not exist", zap.Error(err))
		} else if err != nil {
			handler.Logger.Error("Error while restoring the file", zap.Error(err))
		}
	}

	db, err := sql.Open("pgx", flagDatabaseDSN)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	postgresStorage := repository.NewPostgresStorage(db)

	metricsHandler := handler.MetricsHandler{
		Storage:  persistentMemStorage,
		Database: postgresStorage,
	}
	router := metricsHandler.New()

	srv := &http.Server{
		Addr:         flagHTTPAddr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      router,
	}

	if !isSyncWrite {
		go persistentMemStorage.PeriodicSave(time.Duration(flagStoreInterval) * time.Second)
	}

	handler.Logger.Info("Running server", zap.String("address", flagHTTPAddr), zap.String("logLevel", flagLogLevel))
	return srv.ListenAndServe()
}
