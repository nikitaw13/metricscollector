package main

import (
	"net/http"
	"time"

	"github.com/PrometheRus/metricscollector/internal/agent"
)

func main() {
	parseFlags()
	run()
}

func run() {
	storage := agent.NewAgentStorage()
	collector := &agent.Collector{Storage: storage}
	sender := &agent.Sender{
		URL:     "http://" + flagServerAddr,
		Storage: storage,
		Client: http.Client{
			Timeout: 5 * time.Second,
		},
	}

	// Collector runs in a separate goroutine since two independent intervals
	// cannot be managed by Sleep in a single goroutine.
	go func() {
		for {
			collector.Run()
			time.Sleep(time.Duration(flagPollInterval) * time.Second)
		}
	}()

	for {
		sender.Run()
		time.Sleep(time.Duration(flagReportInterval) * time.Second)
	}
}
