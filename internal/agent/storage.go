package agent

import (
	"fmt"
	"maps"
)

// Интерфейс для взаимодействия с хранилищем
// Единственный потребитель — тесты
type Storage interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, value int64) error
	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)
	GetAllGauges() map[string]float64
	GetAllCounters() map[string]int64
}

// Тип для хранения агентских метрик
type AgentStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

var _ Storage = (*AgentStorage)(nil)

func NewAgentStorage() *AgentStorage {
	return &AgentStorage{
		gauge:   map[string]float64{},
		counter: map[string]int64{},
	}
}

func (ms *AgentStorage) UpdateGauge(name string, value float64) error {
	ms.gauge[name] = value
	return nil
}

func (ms *AgentStorage) UpdateCounter(name string, value int64) error {
	ms.counter[name] += value
	return nil
}

func (ms *AgentStorage) GetGauge(name string) (value float64, err error) {
	if _, ok := ms.gauge[name]; ok {
		return ms.gauge[name], nil
	}
	return 0, fmt.Errorf("Gauge %s not found", name)
}

func (ms *AgentStorage) GetCounter(name string) (value int64, err error) {
	if _, ok := ms.counter[name]; ok {
		return ms.counter[name], nil
	}
	return 0, fmt.Errorf("Counter %s not found", name)
}

func (ms *AgentStorage) GetAllGauges() (m map[string]float64) {
	return maps.Clone(ms.gauge)
}

func (ms *AgentStorage) GetAllCounters() (m map[string]int64) {
	return maps.Clone(ms.counter)
}
