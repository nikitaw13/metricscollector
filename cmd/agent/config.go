package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	// serverAddr holds the HTTP address of the metrics server (-a).
	serverAddr string

	// logLevel holds the agent log level.
	logLevel string

	// reportInterval holds the interval in seconds between metric reports (-r).
	reportInterval int

	// pollInterval holds the interval in seconds between metric collection cycles (-p).
	pollInterval int

	// hashKey is the secret key used for HMAC-SHA256 request signing.
	hashKey string

	// rateLimit is the maximum number of concurrent outgoing requests to the server.
	rateLimit int
}

func LoadConfig() (Config, error) {
	cfg := Config{
		serverAddr:     "localhost:8080",
		logLevel:       "INFO",
		reportInterval: 10,
		pollInterval:   2,
		hashKey:        "",
		rateLimit:      20,
	}

	envAddress, found := os.LookupEnv("ADDRESS")
	if found {
		cfg.serverAddr = envAddress
	}

	envReportInterval, found := os.LookupEnv("REPORT_INTERVAL")
	if found {
		intervalSec, err := strconv.Atoi(envReportInterval)

		if err != nil {
			return cfg, fmt.Errorf("failed to parse REPORT_INTERVAL: %w", err)
		}
		cfg.reportInterval = intervalSec
	}

	envPollInterval, found := os.LookupEnv("POLL_INTERVAL")
	if found {
		intervalSec, err := strconv.Atoi(envPollInterval)

		if err != nil {
			return cfg, fmt.Errorf("failed to parse POLL_INTERVAL: %w", err)
		}

		cfg.pollInterval = intervalSec
	}

	envHashKey, found := os.LookupEnv("KEY")
	if found {
		cfg.hashKey = envHashKey
	}

	envRateLimit, found := os.LookupEnv("RATE_LIMIT")
	if found {
		rateLimit, err := strconv.Atoi(envRateLimit)

		if err != nil {
			return cfg, fmt.Errorf("failed to parse RATE_LIMIT: %w", err)
		}

		cfg.rateLimit = rateLimit
	}

	flag.StringVar(&cfg.serverAddr, "a", cfg.serverAddr, "HTTP server endpoint address")
	flag.IntVar(&cfg.reportInterval, "r", cfg.reportInterval, "Metrics report interval in seconds")
	flag.IntVar(&cfg.pollInterval, "p", cfg.pollInterval, "Metrics poll interval in seconds")
	flag.StringVar(&cfg.hashKey, "k", cfg.hashKey, "secret key for HMAC-SHA256 body signing")
	flag.IntVar(&cfg.rateLimit, "l", cfg.rateLimit, "maximum number of concurrent outgoing requests")
	flag.Parse()

	err := cfg.validateConfig()
	if err != nil {
		return cfg, fmt.Errorf("validate config values: %w", err)
	}

	return cfg, nil
}

func (cfg Config) validateConfig() error {
	if cfg.reportInterval <= 0 {
		return fmt.Errorf("metrics report interval in seconds must be positive: %d", cfg.reportInterval)
	}

	if cfg.pollInterval <= 0 {
		return fmt.Errorf("metrics poll interval in seconds must be positive: %d", cfg.pollInterval)
	}

	if cfg.rateLimit <= 0 {
		return fmt.Errorf("maximum number of concurrent outgoing requests must be positive: %d", cfg.rateLimit)
	}
	return nil
}
