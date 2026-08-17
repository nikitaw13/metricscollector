package main

import (
	"flag"
)

// flagHTTPAddr holds the HTTP listen address.
var flagHTTPAddr string

// flagLogLevel holds the server log level
var flagLogLevel string

// parseFlags registers and parses command-line flags.
func parseFlags() {
	flag.StringVar(&flagHTTPAddr, "a", "localhost:8080", "HTTP server endpoint address")
	flag.StringVar(&flagLogLevel, "l", "Info", "HTTP server log level")
	flag.Parse()
}
