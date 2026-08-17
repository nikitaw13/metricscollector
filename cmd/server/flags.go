package main

import (
	"flag"
)

// Default server endpoint address.
var flagRunAddr string

// Default server log level.
var flagLogLevel string

// parseFlags registers and parses server CLI flags.
func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "HTTP server endpoint address")
	flag.StringVar(&flagLogLevel, "l", "Info", "HTTP server log level")
	flag.Parse()
}
