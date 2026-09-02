package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/PrometheRus/metricscollector/internal/model"
)

// PersistentMemStorage wraps MemStorage with file-based persistence capabilities.
type PersistentMemStorage struct {
	*MemStorage
	filePath  string
	syncWrite bool
}

// NewPersistentMemStorage creates a PersistentMemStorage with the given underlying storage, file path, and sync-write mode.
func NewPersistentMemStorage(memStorage *MemStorage, filePath string, syncWrite bool) *PersistentMemStorage {
	return &PersistentMemStorage{
		MemStorage: memStorage,
		filePath:   filePath,
		syncWrite:  syncWrite,
	}
}

// Restore loads previously saved metrics from the file into in-memory maps.
func (s *PersistentMemStorage) Restore() error {
	file, err := os.Open(s.filePath)

	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}

	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	var metrics []model.Metric
	err = json.Unmarshal(data, &metrics)
	if err != nil {
		return fmt.Errorf("error unmarshaling metrics %v, %w", data, err)
	}

	clear(s.gauge)
	clear(s.counter)

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

// Save writes all current metrics to the file, replacing any previous contents.
func (s *PersistentMemStorage) Save() error {
	file, err := os.OpenFile(s.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)

	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}

	defer file.Close()

	gaugesMap, _ := s.GetAllGauges()
	countersMap, _ := s.GetAllCounters()
	var metrics []model.Metric

	for key, value := range gaugesMap {
		metric := model.Metric{ID: key, Type: "gauge", Value: &value}
		metrics = append(metrics, metric)
	}

	for key, value := range countersMap {
		metric := model.Metric{ID: key, Type: "counter", Delta: &value}
		metrics = append(metrics, metric)
	}

	data, err := json.Marshal(&metrics)
	if err != nil {
		return fmt.Errorf("error marshaling data: %w", err)
	}

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

// PeriodicSave saves metrics to disk at the given interval until the program exits.
func (s *PersistentMemStorage) PeriodicSave(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		err := s.Save()
		if err != nil {
			log.Printf("error writing file: %v", err)
			continue
		}
	}
}

// SaveSync performs a synchronous save and logs any error.
func (s *PersistentMemStorage) SaveSync() {
	err := s.Save()
	if err != nil {
		log.Printf("error writing file: %v", err)
	}
}

// SetGauge sets the named gauge metric to the specified value, overwriting any previous value.
func (s *PersistentMemStorage) SetGauge(name string, value float64) error {
	s.gauge[name] = value
	if s.syncWrite {
		s.SaveSync()
	}
	return nil
}

// AddCounter increments the named counter metric by the specified delta.
func (s *PersistentMemStorage) AddCounter(name string, value int64) error {
	s.counter[name] += value
	if s.syncWrite {
		s.SaveSync()
	}
	return nil
}
