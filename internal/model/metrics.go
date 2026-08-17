package model

import (
	"errors"
	"math"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
)

var (
	ErrMissingType = errors.New("Metric type is required")
	ErrInvalidType = errors.New("Invalid metric type")
	ErrMissingID   = errors.New("Metric ID is required")
	// ErrInvalidID    = errors.New("Invalid ID")
	ErrMissingValue = errors.New("Metric value is required")
	ErrInvalidValue = errors.New("Invalid metric value")
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	Type string   `json:"type"`            // параметр, принимающий значение: gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`
}

func (m Metric) ValidateForUpdate() error {
	if m.Type == "" {
		return ErrMissingType
	}
	if m.Type != Gauge && m.Type != Counter {
		return ErrInvalidType
	}
	if m.ID == "" {
		return ErrMissingID
	}

	if (m.Type == Gauge && m.Value == nil) || (m.Type == Counter && m.Delta == nil) {
		return ErrMissingValue
	}

	if m.Type == Gauge && m.Value != nil && math.IsInf(*m.Value, 0) {
		return ErrInvalidValue
	}
	if m.Type == Gauge && m.Value != nil && math.IsNaN(*m.Value) {
		return ErrInvalidValue
	}

	if m.Type == Counter && m.Delta != nil && *m.Delta < 0 {
		return ErrInvalidValue
	}

	return nil
}

func (m Metric) ValidateForRead() error {
	if m.Type == "" {
		return ErrMissingType
	}
	if m.Type != Gauge && m.Type != Counter {
		return ErrInvalidType
	}
	if m.ID == "" {
		return ErrMissingID
	}
	return nil
}
