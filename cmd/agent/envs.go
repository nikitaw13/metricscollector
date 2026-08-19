package main

import (
	"log"
	"os"
	"strconv"
)

// parseEnvs overrides flag variables with values from environment variables.
func parseEnvs() {
	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		flagServerAddr = envAddress
	}

	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		result, err := strconv.Atoi(envReportInterval)

		if err != nil {
			log.Fatal("Can't convert REPORT_INTERVAL")
		}
		flagReportInterval = result
	}

	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		result, err := strconv.Atoi(envPollInterval)

		if err != nil {
			log.Fatal("Can't convert POLL_INTERVAL")
		}

		flagPollInterval = result
	}
}
