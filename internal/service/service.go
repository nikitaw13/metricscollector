package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/nikitaw13/metricscollector/internal/model"
	"go.uber.org/zap"
)

type Service struct {
	storage  Repository
	database DBPinger
	logger   zap.Logger
}

func New(storage Repository, database DBPinger, logger zap.Logger) *Service {
	return &Service{
		storage:  storage,
		database: database,
		logger:   logger,
	}
}

func (s *Service) ListAllMetrics() (map[string]float64, map[string]int64, error) {
	counters, err := s.storage.GetAllCounters()
	if err != nil {
		return nil, nil, fmt.Errorf("get all counters: %w", err)
	}
	gauges, err := s.storage.GetAllGauges()
	if err != nil {
		return nil, nil, fmt.Errorf("get all gauges: %w", err)
	}
	return gauges, counters, nil
}

func (s *Service) GetCounter(metricName string) (int64, error) {
	value, err := s.storage.GetCounter(metricName)
	if errors.Is(err, model.ErrMetricNotFound) {
		return 0, model.ErrMetricNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("get a counter: %w", err)
	}
	return value, nil

}

func (s *Service) GetGauge(metricName string) (float64, error) {
	value, err := s.storage.GetGauge(metricName)
	if errors.Is(err, model.ErrMetricNotFound) {
		return 0, model.ErrMetricNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("get a gauge: %w", err)
	}
	return value, nil
}

func (s *Service) SetGauge(name string, value float64) error {
	err := s.storage.SetGauge(name, value)
	if err != nil {
		return fmt.Errorf("set a gauge: %w", err)
	}
	return nil
}

func (s *Service) AddCounter(name string, delta int64) (int64, error) {
	newValue, err := s.storage.AddCounter(name, delta)
	if err != nil {
		return newValue, fmt.Errorf("add counter: %w", err)
	}
	return newValue, nil
}

func (s *Service) UpdateMetricsBatch(ctx context.Context, metrics []model.Metric) error {
	err := s.storage.UpdateMetrics(ctx, metrics)
	if err != nil {
		return fmt.Errorf("update batch: %w", err)
	}
	return nil
}
