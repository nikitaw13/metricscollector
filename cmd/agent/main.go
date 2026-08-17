package main

import (
	"net/http"
	"time"

	"github.com/PrometheRus/metricscollector/internal/agent"
)

func main() {
	parseFlags()
	parseEnvs()
	run()
}

func run() {
	storage := agent.NewAgentStorage()
	collector := &agent.Collector{Storage: storage}
	sender := &agent.Sender{
		URL:     "http://" + flagRunAddr,
		Storage: storage,
		Client: http.Client{
			Timeout: 5 * time.Second,
		},
	}

	// Two independent intervals cannot be implemented in a single goroutine with Sleep.
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
