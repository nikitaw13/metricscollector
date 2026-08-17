package main

import (
	"flag"
)

// Default server endpoint address
var flagHTTPAddr string

// Function registers and parses command-line flags
func parseFlags() {
	flag.StringVar(&flagHTTPAddr, "a", "localhost:8080", "HTTP server endpoint address")
	flag.Parse()
}
