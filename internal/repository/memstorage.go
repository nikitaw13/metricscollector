package repository

import (
	"fmt"
	"maps"
)

// Struct implements in-memory metric storage using separate maps for gauge and counter metrics.
type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

var _ Repository = (*MemStorage)(nil)

// Function creates and returns a pointer to an initialized MemStorage instance with empty metric maps.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauge:   map[string]float64{},
		counter: map[string]int64{},
	}
}

// Method sets the gauge metric identified by name to the specified value, overwriting any previous value.
func (ms *MemStorage) UpdateGauge(name string, value float64) error {
	ms.gauge[name] = value
	return nil
}

// Method increments the counter metric identified by name by the specified delta value.
func (ms *MemStorage) UpdateCounter(name string, value int64) error {
	ms.counter[name] += value
	return nil
}

// Method returns the value of the gauge metric identified by name.
// Returns an error if the metric does not exist.
func (ms *MemStorage) GetGauge(name string) (value float64, err error) {
	if _, ok := ms.gauge[name]; ok {
		return ms.gauge[name], nil
	}
	return 0, fmt.Errorf("Gauge %s not found", name)
}

// Method returns the value of the counter metric identified by name.
// Returns an error if the metric does not exist.
func (ms *MemStorage) GetCounter(name string) (value int64, err error) {
	if _, ok := ms.counter[name]; ok {
		return ms.counter[name], nil
	}
	return 0, fmt.Errorf("Counter %s not found", name)
}

// Method returns a shallow copy of all stored gauge metrics to prevent external mutation of internal state.
func (ms *MemStorage) GetAllGauges() (m map[string]float64) {
	return maps.Clone(ms.gauge)
}

// Method returns a shallow copy of all stored counter metrics to prevent external mutation of internal state.
func (ms *MemStorage) GetAllCounters() (m map[string]int64) {
	return maps.Clone(ms.counter)
}
