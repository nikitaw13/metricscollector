package main

import (
	"os"
)

func parseEnvs() {
	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		flagHTTPAddr = envAddress
	}
}
