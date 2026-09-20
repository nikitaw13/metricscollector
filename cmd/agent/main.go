package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/nikitaw13/metricscollector/internal/agent"
	"github.com/nikitaw13/metricscollector/internal/model"
)

func main() {
	parseFlags()
	parseEnvs()
	run()
}

// worker consumes metric batches from the jobs channel and sends them to the server.
func worker(sender *agent.Sender, jobs <-chan []model.Metric) {
	for batch := range jobs {
		sender.Run(batch)
	}
}

func run() {
	var (
		baseURL     = fmt.Sprintf("http://%s", flagServerAddr)
		timeouts    = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
		httpClient  = &http.Client{Timeout: 5 * time.Second}
		retryClient = agent.NewClientWithRetries(timeouts, httpClient)
		sender      = agent.NewSender(baseURL, retryClient, flagHashKey)
	)

	jobs := make(chan []model.Metric, flagRateLimit)

	for workerNum := 1; workerNum <= flagRateLimit; workerNum++ {
		go worker(sender, jobs)
	}

	snapshot := make(chan []model.Metric, 1)

	// The dispatch goroutine forwards the latest snapshot to the jobs channel once per report interval.
	go func() {
		ticker := time.NewTicker(time.Duration(flagReportInterval) * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			select {
			case batch := <-snapshot:
				jobs <- batch
			default:
			}
		}
	}()

	// The collector goroutine gathers one snapshot per poll interval, keeping only the latest one in the snapshot channel.
	go func() {
		ticker := time.NewTicker(time.Duration(flagPollInterval) * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			batch := append(agent.CollectRuntimeMetrics(), agent.CollectSystemMetrics()...)
			select {
			// The snapshot channel has room: offer the freshly collected batch.
			case snapshot <- batch:
			// The channel still holds the previous snapshot: drop it and put in the fresh one.
			default:
				<-snapshot
				snapshot <- batch
			}
		}
	}()

	// Block forever; collection and sending run in their own goroutines.
	select {}
}
