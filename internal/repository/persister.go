package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/PrometheRus/metricscollector/internal/model"
)

type PersistentMemStorage struct {
	*MemStorage
	FilePath   string
	Synchronic bool
	//mu       sync.Mutex
}

func NewPersistentMemStorage(ms *MemStorage, fp string, s bool) *PersistentMemStorage {
	return &PersistentMemStorage{
		MemStorage: ms,
		FilePath:   fp,
		Synchronic: s,
	}
}

func (s *PersistentMemStorage) Restore() error {
	file, err := os.Open(s.FilePath)

	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("Error opening file: %w", err)
	}

	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("Error reading file: %w", err)
	}

	var metrics []model.Metric
	err = json.Unmarshal(data, &metrics)
	if err != nil {
		return fmt.Errorf("Error unmarshaling metrics %v, %w", data, err)
	}

	for _, m := range metrics {
		switch m.Type {
		case "gauge":
			s.gauge[m.ID] = *m.Value
		case "counter":
			s.counter[m.ID] = *m.Delta
		}
	}

	return nil

}

func (s *PersistentMemStorage) PeriodicSave(interval int) {

}

func (s *PersistentMemStorage) SynchronicSave() {

}

// SetGauge sets the named gauge metric to the specified value, overwriting any previous value.
func (s *PersistentMemStorage) SetGauge(name string, value float64) error {
	s.gauge[name] = value
	if s.Synchronic {
		s.SynchronicSave()
	}
	return nil
}

// AddCounter increments the named counter metric by the specified delta.
func (s *PersistentMemStorage) AddCounter(name string, value int64) error {
	s.counter[name] += value
	if s.Synchronic {
		s.SynchronicSave()
	}
	return nil
}
