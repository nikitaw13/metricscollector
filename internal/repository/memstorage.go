package repository

import (
	"fmt"
	"maps"
)

// MemStorage implements in-memory metric storage using separate maps for gauge and counter metrics.
type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

// New creates and returns an initialized MemStorage with empty metric maps.
func New() *MemStorage {
	return &MemStorage{
		gauge:   map[string]float64{},
		counter: map[string]int64{},
	}
}

// SetGauge sets the named gauge metric to the specified value, overwriting any previous value.
func (ms *MemStorage) SetGauge(name string, value float64) error {
	ms.gauge[name] = value
	return nil
}

// AddCounter increments the named counter metric by the specified delta.
func (ms *MemStorage) AddCounter(name string, value int64) error {
	ms.counter[name] += value
	return nil
}

// GetGauge returns the value of the named gauge metric.
// Returns an error if the metric does not exist.
func (ms *MemStorage) GetGauge(name string) (value float64, err error) {
	if _, ok := ms.gauge[name]; ok {
		return ms.gauge[name], nil
	}
	return 0, fmt.Errorf("Gauge %s not found", name)
}

// GetCounter returns the value of the named counter metric.
// Returns an error if the metric does not exist.
func (ms *MemStorage) GetCounter(name string) (value int64, err error) {
	if _, ok := ms.counter[name]; ok {
		return ms.counter[name], nil
	}
	return 0, fmt.Errorf("Counter %s not found", name)
}

// GetAllGauges returns a shallow copy of all gauge metrics to prevent external mutation.
func (ms *MemStorage) GetAllGauges() (m map[string]float64) {
	return maps.Clone(ms.gauge)
}

// GetAllCounters returns a shallow copy of all counter metrics to prevent external mutation.
func (ms *MemStorage) GetAllCounters() (m map[string]int64) {
	return maps.Clone(ms.counter)
}
