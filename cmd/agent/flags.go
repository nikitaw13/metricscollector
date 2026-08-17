package main

import (
	"flag"
)

// Default agent configuration values.
var (
	flagRunAddr        string
	flagReportInterval int
	flagPollInterval   int
)

// parseFlags registers and parses agent CLI flags.
func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "HTTP server endpoint address")
	flag.IntVar(&flagReportInterval, "r", 10, "Metrics report interval in seconds")
	flag.IntVar(&flagPollInterval, "p", 2, "Metrics poll interval in seconds")
	flag.Parse()
}