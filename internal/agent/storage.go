package agent

import (
	"fmt"
	"maps"
	"sync"

	"github.com/nikitaw13/metricscollector/internal/model"
)

// AgentStorage holds agent metrics in memory.
type AgentStorage struct {
	mu      sync.Mutex
	gauge   map[string]float64
	counter map[string]int64
}

// NewAgentStorage creates and returns a new AgentStorage with initialized maps.
func NewAgentStorage() *AgentStorage {
	return &AgentStorage{
		gauge:   map[string]float64{},
		counter: map[string]int64{},
	}
}

// SetGauge sets the gauge metric with the given name to the specified value.
func (s *AgentStorage) SetGauge(name string, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.gauge[name] = value
}

// AddCounter increments the counter metric with the given name by the specified value.
func (s *AgentStorage) AddCounter(name string, value int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter[name] += value
}

// GetGauge returns the value of the gauge metric by name.
// Returns an error if the metric does not exist.
func (s *AgentStorage) GetGauge(name string) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.gauge[name]; ok {
		return s.gauge[name], nil
	}
	return 0, fmt.Errorf("gauge %s %w", name, model.ErrMetricNotFound)
}

// GetCounter returns the value of the counter metric by name.
// Returns an error if the metric does not exist.
func (s *AgentStorage) GetCounter(name string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.counter[name]; ok {
		return s.counter[name], nil
	}
	return 0, fmt.Errorf("counter %s %w", name, model.ErrMetricNotFound)
}

// GetAllGauges returns a copy of all gauge metrics to prevent external mutation.
func (s *AgentStorage) GetAllGauges() map[string]float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	return maps.Clone(s.gauge)
}

// GetAllCounters returns a copy of all counter metrics to prevent external mutation.
func (s *AgentStorage) GetAllCounters() map[string]int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	return maps.Clone(s.counter)
}

// DrainCounters atomically returns current counter values and resets them to zero.
func (s *AgentStorage) DrainCounters() map[string]int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	snapshot := maps.Clone(s.counter)
	clear(s.counter)
	return snapshot
}
