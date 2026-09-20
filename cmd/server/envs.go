package main

import (
	"fmt"
	"os"
	"strconv"
)

// parseEnvs overrides flag variables with values from environment variables.
// It returns an error if a variable value cannot be parsed.
func parseEnvs() error {
	envAddress, found := os.LookupEnv("ADDRESS")
	if found {
		flagHTTPAddr = envAddress
	}

	envLogLevel, found := os.LookupEnv("LOG_LEVEL")
	if found {
		flagLogLevel = envLogLevel
	}

	envStoreInterval, found := os.LookupEnv("STORE_INTERVAL")
	if found {
		intervalSec, err := strconv.Atoi(envStoreInterval)

		if err != nil {
			return fmt.Errorf("failed to parse STORE_INTERVAL: %w", err)
		}
		flagStoreInterval = intervalSec
	}

	envFileStoragePath, found := os.LookupEnv("FILE_STORAGE_PATH")
	if found {
		flagFileStoragePath = envFileStoragePath
	}

	envRestore, found := os.LookupEnv("RESTORE")
	if found {
		shouldRestore, err := strconv.ParseBool(envRestore)

		if err != nil {
			return fmt.Errorf("failed to parse RESTORE: %w", err)
		}
		flagRestore = shouldRestore
	}

	envDatabaseDSN, found := os.LookupEnv("DATABASE_DSN")
	if found {
		flagDatabaseDSN = envDatabaseDSN
	}

	envMigrationPath, found := os.LookupEnv("MIGRATION_PATH")
	if found {
		flagMigrationPath = envMigrationPath
	}

	envHashKey, found := os.LookupEnv("KEY")
	if found {
		flagHashKey = envHashKey
	}
	return nil
}
