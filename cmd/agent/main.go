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

// worker consumes metric batches from the jobs channel and sends them to the server, pacing deliveries with the report interval.
func worker(sender *agent.Sender, jobs <-chan []model.Metric) {
	for batch := range jobs {
		sender.Run(batch)
		time.Sleep(time.Duration(flagReportInterval) * time.Second)
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

	for w := 1; w <= flagRateLimit; w++ {
		go worker(sender, jobs)
	}

	go func() {
		for {
			jobs <- agent.Collect()
			time.Sleep(time.Duration(flagPollInterval) * time.Second)
		}
	}()
	// Block forever; collection and sending run in their own goroutines.
	select {}
}
