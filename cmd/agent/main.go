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

	// Два интервала в одном потоке через Sleep не реализовать
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
