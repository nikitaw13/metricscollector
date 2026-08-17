package main

import (
	"os"
)

func parseEnvs() {
	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		flagHTTPAddr = envAddress
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		flagLogLevel = envLogLevel
	}
}
