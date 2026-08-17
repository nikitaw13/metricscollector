package main

import (
	"flag"
)

// flagHTTPAddr holds the HTTP listen address.
var flagHTTPAddr string

// parseFlags registers and parses command-line flags.
func parseFlags() {
	flag.StringVar(&flagHTTPAddr, "a", "localhost:8080", "HTTP server endpoint address")
	flag.Parse()
}
