package main

import (
	"log"
	"os"
	"strconv"
)

// parseEnvs overrides flag variables with values from environment variables.
func parseEnvs() {
	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		flagHTTPAddr = envAddress
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		flagLogLevel = envLogLevel
	}

	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		intervalSec, err := strconv.Atoi(envStoreInterval)

		if err != nil {
			log.Fatal("Can't convert STORE_INTERVAL")
		}
		flagStoreInterval = intervalSec
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		flagFileStoragePath = envFileStoragePath
	}

	if envRESTORE := os.Getenv("RESTORE"); envRESTORE != "" {
		shouldRestore, err := strconv.ParseBool(envRESTORE)

		if err != nil {
			log.Fatal("Can't convert RESTORE")
		}
		flagRestore = shouldRestore
	}
}
