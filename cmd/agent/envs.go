package main

import (
	"log"
	"os"
	"strconv"
)

// parseEnvs overrides flag variables with values from environment variables.
func parseEnvs() {
	envAddress, found := os.LookupEnv("ADDRESS")
	if found {
		flagServerAddr = envAddress
	}

	envReportInterval, found := os.LookupEnv("REPORT_INTERVAL")
	if found {
		intervalSec, err := strconv.Atoi(envReportInterval)

		if err != nil {
			log.Fatalf("failed to parse REPORT_INTERVAL: %v", err)
		}
		flagReportInterval = intervalSec
	}

	envPollInterval, found := os.LookupEnv("POLL_INTERVAL")
	if found {
		intervalSec, err := strconv.Atoi(envPollInterval)

		if err != nil {
			log.Fatalf("failed to parse POLL_INTERVAL: %v", err)
		}

		flagPollInterval = intervalSec
	}

	envHashKey, found := os.LookupEnv("KEY")
	if found {
		flagHashKey = envHashKey
	}

	envRateLimit, found := os.LookupEnv("RATE_LIMIT")
	if found {
		rateLimit, err := strconv.Atoi(envRateLimit)

		if err != nil {
			log.Fatalf("failed to parse RATE_LIMIT: %v", err)
		}

		flagRateLimit = rateLimit
	}
}
