package main

import (
	"flag"
)

// Default server endpoint address
var flagRunAddr string

// Function registers and parses command-line flags
func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "HTTP server endpoint address")
	flag.Parse()
}
