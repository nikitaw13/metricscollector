package agent

import (
	"fmt"
	"maps"
)

// StorageGetter defines read-only operations for retrieving metrics.
// The only consumer is tests.
type StorageGetter interface {
	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)
	GetAllGauges() map[string]float64
	GetAllCounters() map[string]int64
}

// StorageSetter defines write operations for updating metrics.
type StorageSetter interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, value int64) error
}

type CounterReseter interface {
	ResetCounter(name string) error
}

// Storage combines read and write operations for the agent's metric store.
type Storage interface {
	StorageGetter
	StorageSetter
	CounterReseter
}

// AgentStorage is an in-memory implementation of Storage that holds agent metrics.
type AgentStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

// Compile-time check that AgentStorage implements Storage.
var _ Storage = (*AgentStorage)(nil)

// NewAgentStorage creates and returns a new AgentStorage with initialized maps.
func NewAgentStorage() *AgentStorage {
	return &AgentStorage{
		gauge:   map[string]float64{},
		counter: map[string]int64{},
	}
}

// UpdateGauge sets the gauge metric with the given name to the specified value.
func (ms *AgentStorage) UpdateGauge(name string, value float64) error {
	ms.gauge[name] = value
	return nil
}

// UpdateCounter increments the counter metric with the given name by the specified value.
func (ms *AgentStorage) UpdateCounter(name string, value int64) error {
	ms.counter[name] += value
	return nil
}

// ResetCounters reset the counter metric to zero after successful report
func (ms *AgentStorage) ResetCounter(name string) error {
	ms.counter[name] = 0
	return nil
}

// GetGauge returns the value of the gauge metric by name.
// Returns an error if the metric does not exist.
func (ms *AgentStorage) GetGauge(name string) (value float64, err error) {
	if _, ok := ms.gauge[name]; ok {
		return ms.gauge[name], nil
	}
	return 0, fmt.Errorf("Gauge %s not found", name)
}

// GetCounter returns the value of the counter metric by name.
// Returns an error if the metric does not exist.
func (ms *AgentStorage) GetCounter(name string) (value int64, err error) {
	if _, ok := ms.counter[name]; ok {
		return ms.counter[name], nil
	}
	return 0, fmt.Errorf("Counter %s not found", name)
}

// GetAllGauges returns a copy of all gauge metrics to prevent external mutation.
func (ms *AgentStorage) GetAllGauges() (m map[string]float64) {
	return maps.Clone(ms.gauge)
}

// GetAllCounters returns a copy of all counter metrics to prevent external mutation.
func (ms *AgentStorage) GetAllCounters() (m map[string]int64) {
	return maps.Clone(ms.counter)
}
