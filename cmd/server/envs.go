package main

import (
	"os"
)

// parseEnvs overrides flag variables with values from environment variables.
func parseEnvs() {
	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		flagHTTPAddr = envAddress
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		flagLogLevel = envLogLevel
	}
}
