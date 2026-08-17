package repository

import (
	"fmt"
	"maps"
)

// MemStorage holds gauge and counter metrics in separate maps.
type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

// New returns an empty MemStorage with initialized maps.
func New() *MemStorage {
	return &MemStorage{
		gauge:   map[string]float64{},
		counter: map[string]int64{},
	}
}

// UpdateGauge overwrites the gauge metric with the given value.
func (ms *MemStorage) UpdateGauge(name string, value float64) error {
	ms.gauge[name] = value
	return nil
}

// UpdateCounter adds delta to the counter metric.
func (ms *MemStorage) UpdateCounter(name string, value int64) error {
	ms.counter[name] += value
	return nil
}

// GetGauge returns the gauge value or an error if not found.
func (ms *MemStorage) GetGauge(name string) (value float64, err error) {
	if _, ok := ms.gauge[name]; ok {
		return ms.gauge[name], nil
	}
	return 0, fmt.Errorf("Gauge %s not found", name)
}

// GetCounter returns the counter value or an error if not found.
func (ms *MemStorage) GetCounter(name string) (value int64, err error) {
	if _, ok := ms.counter[name]; ok {
		return ms.counter[name], nil
	}
	return 0, fmt.Errorf("Counter %s not found", name)
}

// GetAllGauges returns a defensive copy of all gauge metrics.
func (ms *MemStorage) GetAllGauges() (m map[string]float64) {
	return maps.Clone(ms.gauge)
}

// GetAllCounters returns a defensive copy of all counter metrics.
func (ms *MemStorage) GetAllCounters() (m map[string]int64) {
	return maps.Clone(ms.counter)
}