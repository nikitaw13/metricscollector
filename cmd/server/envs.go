package main

import (
	"os"
)

// parseEnvs overrides CLI flags with environment variables when set.
func parseEnvs() {
	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		flagRunAddr = envAddress
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		flagLogLevel = envLogLevel
	}
}