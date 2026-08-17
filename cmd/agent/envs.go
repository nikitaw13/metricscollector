package main

import (
	"log"
	"os"
	"strconv"
)

// parseEnvs overrides CLI flags with environment variables when set.
func parseEnvs() {
	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		flagRunAddr = envAddress
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
