package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/PrometheRus/metricscollector/internal/model"
)

type PersistentMemStorage struct {
	*MemStorage
	FilePath  string
	SyncWrite bool
}

func NewPersistentMemStorage(memStorage *MemStorage, filePath string, syncWrite bool) *PersistentMemStorage {
	return &PersistentMemStorage{
		MemStorage: memStorage,
		FilePath:   filePath,
		SyncWrite: syncWrite,
	}
}

func (s *PersistentMemStorage) Restore() error {
	file, err := os.Open(s.FilePath)

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

func (s *PersistentMemStorage) Save() error {
	file, err := os.OpenFile(s.FilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)

	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}

	defer file.Close()

	gaugesMap := s.GetAllGauges()
	countersMap := s.GetAllCounters()
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

func (s *PersistentMemStorage) PeriodicSave(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		err := s.Save()
		if err != nil {
			fmt.Printf("error writing file: %v", err)
			continue
		}
	}
}

func (s *PersistentMemStorage) SaveSync() {
	err := s.Save()
	if err != nil {
		fmt.Printf("error writing file: %v", err)
	}

}

// SetGauge sets the named gauge metric to the specified value, overwriting any previous value.
func (s *PersistentMemStorage) SetGauge(name string, value float64) error {
	s.gauge[name] = value
	if s.SyncWrite {
		s.SaveSync()
	}
	return nil
}

// AddCounter increments the named counter metric by the specified delta.
func (s *PersistentMemStorage) AddCounter(name string, value int64) error {
	s.counter[name] += value
	if s.SyncWrite {
		s.SaveSync()
	}
	return nil
}
