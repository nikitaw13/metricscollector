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
		flagServerAddr = envAddress
	}

	envReportInterval, found := os.LookupEnv("REPORT_INTERVAL")
	if found {
		intervalSec, err := strconv.Atoi(envReportInterval)

		if err != nil {
			return fmt.Errorf("failed to parse REPORT_INTERVAL: %w", err)
		}
		flagReportInterval = intervalSec
	}

	envPollInterval, found := os.LookupEnv("POLL_INTERVAL")
	if found {
		intervalSec, err := strconv.Atoi(envPollInterval)

		if err != nil {
			return fmt.Errorf("failed to parse POLL_INTERVAL: %w", err)
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
			return fmt.Errorf("failed to parse RATE_LIMIT: %w", err)
		}

		flagRateLimit = rateLimit
	}
	return nil
}
