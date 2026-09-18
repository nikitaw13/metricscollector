package main

import (
	"flag"
)

// flagServerAddr holds the HTTP address of the metrics server (-a).
var flagServerAddr string

// flagReportInterval holds the interval in seconds between metric reports (-r).
var flagReportInterval int

// flagPollInterval holds the interval in seconds between metric collection cycles (-p).
var flagPollInterval int

// flagHashKey is the secret key used for HMAC-SHA256 request signing.
var flagHashKey string

// flagRateLimit is the maximum number of concurrent outgoing requests to the server.
var flagRateLimit int

// parseFlags registers and parses command-line flags.
func parseFlags() {
	flag.StringVar(&flagServerAddr, "a", "localhost:8080", "HTTP server endpoint address")
	flag.IntVar(&flagReportInterval, "r", 10, "Metrics report interval in seconds")
	flag.IntVar(&flagPollInterval, "p", 2, "Metrics poll interval in seconds")
	flag.StringVar(&flagHashKey, "k", "", "secret key for HMAC-SHA256 body signing")
	flag.IntVar(&flagRateLimit, "l", 20, "maximum number of concurrent outgoing requests")
	flag.Parse()
}
