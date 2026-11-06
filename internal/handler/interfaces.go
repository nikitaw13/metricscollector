package handler

import (
	"context"

	"github.com/nikitaw13/metricscollector/internal/model"
)

// Service defines the metric operations the handler layer depends on.
type Service interface {
	ListAllMetrics() (map[string]float64, map[string]int64, error)
	GetCounter(name string) (int64, error)
	GetGauge(name string) (float64, error)
	SetGauge(name string, value float64) error
	AddCounter(name string, delta int64) (int64, error)
	UpdateMetricsBatch(ctx context.Context, metrics []model.Metric) error
	DatabasePing(ctx context.Context) error
}

// Auditor publishes audit events for accepted metric updates,
// recording the request timestamp, metric names and the client IP address.
type Auditor interface {
	Notify(timestamp int64, metrics []model.Metric, ipAddress string) error
}
