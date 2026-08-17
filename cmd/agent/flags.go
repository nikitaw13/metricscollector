package main

import (
	"flag"
)

// Command-line flags for the agent.
var (
	flagServerAddr     string
	flagReportInterval int
	flagPollInterval   int
)

// parseFlags registers and parses command-line flags.
func parseFlags() {
	flag.StringVar(&flagServerAddr, "a", "localhost:8080", "HTTP server endpoint address")
	flag.IntVar(&flagReportInterval, "r", 10, "Metrics report interval in seconds")
	flag.IntVar(&flagPollInterval, "p", 2, "Metrics poll interval in seconds")
	flag.Parse()
}
