package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nikitaw13/metricscollector/internal/agent"
	"github.com/nikitaw13/metricscollector/internal/model"
)

func main() {
	parseFlags()
	if err := parseEnvs(); err != nil {
		log.Fatal(err)
	}
	if err := validateFlags(); err != nil {
		log.Fatal(err)
	}
	run()
}

// worker consumes metric batches from the jobs channel and sends them to the server.
func worker(sender *agent.Sender, jobs <-chan []model.Metric) {
	for batch := range jobs {
		sender.SendBatch(batch)
	}
}

// dispatchSnapshots forwards the latest snapshot to the jobs channel once per report interval until ctx is cancelled.
func dispatchSnapshots(ctx context.Context, snapshot <-chan []model.Metric, jobs chan<- []model.Metric, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		// Wait for the next tick or cancellation.
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Check whether a fresh snapshot is available.
			select {
			case batch := <-snapshot:
				// Send the batch or bail out on cancellation.
				select {
				case jobs <- batch:
				case <-ctx.Done():
					return
				}
			default:
			}
		}
	}
}

// collectMetrics gathers a fresh snapshot of metrics every poll interval,
// keeping only the latest one in the snapshot channel until ctx is cancelled.
func collectMetrics(ctx context.Context, snapshot chan []model.Metric, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		// Wait for the next tick or cancellation.
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			batch := append(agent.CollectRuntimeMetrics(), agent.CollectSystemMetrics()...)
			select {
			case snapshot <- batch: // The slot is free: offer the fresh batch.
			default: // The slot is taken.
				select {
				case <-snapshot: // Dropped a stale batch.
				default: // The dispatch goroutine just drained it: the slot is already empty.
				}
				snapshot <- batch // The slot is now free.
			}
		}
	}
}

// run starts the metric pipeline and blocks until SIGINT or SIGTERM is received.
func run() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var producersWg, workersWg sync.WaitGroup

	var (
		baseURL     = fmt.Sprintf("http://%s", flagServerAddr)
		timeouts    = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
		httpClient  = &http.Client{Timeout: 5 * time.Second}
		retryClient = agent.NewClientWithRetries(timeouts, httpClient)
		sender      = agent.NewSender(baseURL, retryClient, flagHashKey)
	)

	jobs := make(chan []model.Metric, flagRateLimit)

	for workerNum := 1; workerNum <= flagRateLimit; workerNum++ {
		workersWg.Go(func() {
			worker(sender, jobs)
		})
	}

	snapshot := make(chan []model.Metric, 1)

	producersWg.Go(func() {
		dispatchSnapshots(ctx, snapshot, jobs, time.Duration(flagReportInterval)*time.Second)
	})

	producersWg.Go(func() {
		collectMetrics(ctx, snapshot, time.Duration(flagPollInterval)*time.Second)
	})

	<-ctx.Done()

	// Wait for the producers, then close jobs so the workers drain the remaining batches and exit.
	producersWg.Wait()
	close(jobs)
	workersWg.Wait()
}
