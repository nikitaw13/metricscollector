package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	// serverHTTPAddr holds the HTTP listen address.
	serverHTTPAddr string

	// logLevel holds the server log level.
	logLevel string

	// storeInterval is the interval in seconds between periodic metric saves to disk; 0 enables synchronous writes.
	storeInterval int

	// fileStoragePath is the file path where metric values are persisted.
	fileStoragePath string

	// restore determines whether previously saved metrics are loaded from file on server startup.
	restore bool

	// databaseDSN holds the PostgreSQL connection string (DSN).
	databaseDSN string

	// migrationPath is the directory containing database migration files.
	migrationPath string

	// hashKey is the secret key used for HMAC-SHA256 request and response signing.
	hashKey string

	// auditFile is the path to the audit log file.
	auditFile string

	// auditURL is the endpoint that receives audit events.
	auditURL string
}

func LoadConfig() (Config, error) {
	cfg := Config{
		serverHTTPAddr:  "localhost:8080",
		logLevel:        "INFO",
		storeInterval:   300,
		fileStoragePath: "",
		restore:         true,
		databaseDSN:     "",
		migrationPath:   "migrations",
		hashKey:         "",
		auditFile:       "",
		auditURL:        "",
	}

	envAddress, found := os.LookupEnv("ADDRESS")
	if found {
		cfg.serverHTTPAddr = envAddress
	}

	envLogLevel, found := os.LookupEnv("LOG_LEVEL")
	if found {
		cfg.logLevel = envLogLevel
	}

	envStoreInterval, found := os.LookupEnv("STORE_INTERVAL")
	if found {
		intervalSec, err := strconv.Atoi(envStoreInterval)

		if err != nil {
			return cfg, fmt.Errorf("failed to parse STORE_INTERVAL: %w", err)
		}
		cfg.storeInterval = intervalSec
	}

	envFileStoragePath, found := os.LookupEnv("FILE_STORAGE_PATH")
	if found {
		cfg.fileStoragePath = envFileStoragePath
	}

	envRestore, found := os.LookupEnv("RESTORE")
	if found {
		shouldRestore, err := strconv.ParseBool(envRestore)

		if err != nil {
			return cfg, fmt.Errorf("failed to parse RESTORE: %w", err)
		}
		cfg.restore = shouldRestore
	}

	envDatabaseDSN, found := os.LookupEnv("DATABASE_DSN")
	if found {
		cfg.databaseDSN = envDatabaseDSN
	}

	envMigrationPath, found := os.LookupEnv("MIGRATION_PATH")
	if found {
		cfg.migrationPath = envMigrationPath
	}

	envHashKey, found := os.LookupEnv("KEY")
	if found {
		cfg.hashKey = envHashKey
	}

	envAuditFile, found := os.LookupEnv("AUDIT_FILE")
	if found {
		cfg.auditFile = envAuditFile
	}

	envAuditURL, found := os.LookupEnv("AUDIT_URL")
	if found {
		cfg.auditURL = envAuditURL
	}

	flag.StringVar(&cfg.serverHTTPAddr, "a", cfg.serverHTTPAddr, "HTTP server endpoint address")
	flag.StringVar(&cfg.logLevel, "l", cfg.logLevel, "HTTP server log level")
	flag.IntVar(&cfg.storeInterval, "i", cfg.storeInterval, "interval in seconds between periodic saves to disk; 0 makes writes synchronous")
	flag.StringVar(&cfg.fileStoragePath, "f", cfg.fileStoragePath, "file path for persisting current metric values")
	flag.BoolVar(&cfg.restore, "r", cfg.restore, "load previously saved metric values from file on startup")
	flag.StringVar(&cfg.databaseDSN, "d", cfg.databaseDSN, "Database connection DSN")
	flag.StringVar(&cfg.migrationPath, "m", cfg.migrationPath, "directory containing database migration files")
	flag.StringVar(&cfg.hashKey, "k", cfg.hashKey, "secret key for HMAC-SHA256 body signing")
	flag.StringVar(&cfg.auditFile, "audit-file", cfg.auditFile, "path to the audit log file")
	flag.StringVar(&cfg.auditURL, "audit-url", cfg.auditURL, "endpoint that receives audit events")
	flag.Parse()

	err := cfg.Validate()
	if err != nil {
		return cfg, fmt.Errorf("validate config values: %w", err)
	}

	return cfg, nil
}

func (cfg Config) Validate() error {
	if cfg.storeInterval < 0 {
		return fmt.Errorf("interval in seconds between periodic saves to disk must not be negative: %d", cfg.storeInterval)
	}
	return nil
}
