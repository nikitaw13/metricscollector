package repository

import (
	"context"
	"fmt"
	"log"
	"maps"
	"sync"

	"github.com/PrometheRus/metricscollector/internal/model"
)

// MemStorage implements in-memory metric storage using separate maps for gauge and counter metrics.
type MemStorage struct {
	mu      sync.RWMutex
	gauge   map[string]float64
	counter map[string]int64
}

// NewMemStorage creates and returns an initialized MemStorage with empty metric maps.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauge:   map[string]float64{},
		counter: map[string]int64{},
	}
}

// SetGauge sets the named gauge metric to the specified value, overwriting any previous value.
func (ms *MemStorage) SetGauge(name string, value float64) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.gauge[name] = value
	return nil
}

// AddCounter increments the named counter metric by the specified delta.
func (ms *MemStorage) AddCounter(name string, value int64) (newDelta int64, err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.counter[name] += value
	return ms.counter[name], nil
}

// GetGauge returns the value of the named gauge metric.
// Returns an error if the metric does not exist.
func (ms *MemStorage) GetGauge(name string) (value float64, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if _, ok := ms.gauge[name]; ok {
		return ms.gauge[name], nil
	}
	return 0, fmt.Errorf("gauge %s %w", name, ErrMetricNotFound)
}

// GetCounter returns the value of the named counter metric.
// Returns an error if the metric does not exist.
func (ms *MemStorage) GetCounter(name string) (value int64, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if _, ok := ms.counter[name]; ok {
		return ms.counter[name], nil
	}
	return 0, fmt.Errorf("counter %s %w", name, ErrMetricNotFound)
}

// GetAllGauges returns a shallow copy of all gauge metrics to prevent external mutation.
func (ms *MemStorage) GetAllGauges() (result map[string]float64, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return maps.Clone(ms.gauge), nil
}

// GetAllCounters returns a shallow copy of all counter metrics to prevent external mutation.
func (ms *MemStorage) GetAllCounters() (result map[string]int64, err error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return maps.Clone(ms.counter), nil
}

func (ms *MemStorage) UpdateMetrics(ctx context.Context, metrics []model.Metric) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for _, metric := range metrics {
		switch metric.Type {
		case model.Counter:
			ms.counter[metric.ID] += *metric.Delta
		case model.Gauge:
			ms.gauge[metric.ID] = *metric.Value
		default:
			log.Printf("unknown metric type: %s\n", metric.Type)
		}
	}

	return nil
}

func (ms *MemStorage) LoadMetrics(metrics []model.Metric) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	clear(ms.gauge)
	clear(ms.counter)

	for _, metric := range metrics {
		switch metric.Type {
		case model.Counter:
			ms.counter[metric.ID] = *metric.Delta
		case model.Gauge:
			ms.gauge[metric.ID] = *metric.Value
		default:
			log.Printf("unknown metric type: %s\n", metric.Type)
		}

	}

	return nil
}
