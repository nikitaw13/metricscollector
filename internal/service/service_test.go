package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/nikitaw13/metricscollector/internal/mocks"
	"github.com/nikitaw13/metricscollector/internal/model"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

var (
	errTestGetAllCounters = errors.New("failed to get a counter map")
	errTestGetAllGauges   = errors.New("failed to get a gauge map")
	errTestGetCounter     = errors.New("failed to get a counter")
	errTestGetGauge       = errors.New("failed to get a gauge")
	errTestAddCounter     = errors.New("failed to add a counter")
	errTestSetGauge       = errors.New("failed to set a gauge")
	errTestUpdateMetrics  = errors.New("failed to update metrics")
)

var (
	dbPinger              DBPinger
	testGaugeMetricName   = "testGaugeMetric"
	testCounterMetricName = "testCounterMetric"
	mockGaugeValue        = float64(1.00)
	mockCounterValue      = int64(1)
	mockMetricsBatch      = []model.Metric{
		model.Metric{
			ID:    testCounterMetricName,
			Type:  model.Counter,
			Delta: &mockCounterValue,
		},
		model.Metric{
			ID:    testGaugeMetricName,
			Type:  model.Gauge,
			Value: &mockGaugeValue,
		},
	}
	mockGaugeMap     = map[string]float64{"test_gauge": mockGaugeValue}
	mockCounterMap   = map[string]int64{"test_counter": mockCounterValue}
	wantGaugeMap     = map[string]float64{"test_gauge": mockGaugeValue}
	wantCounterMap   = map[string]int64{"test_counter": mockCounterValue}
	wantGaugeValue   = float64(1.00)
	wantCounterValue = int64(1)
)

func TestService_ListAllMetrics(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(repo *mocks.MockRepository)
		wantGauges   map[string]float64
		wantCounters map[string]int64
		wantErr      error
	}{
		{
			name: "return both maps",
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetAllCounters().Return(mockCounterMap, nil)
				repo.EXPECT().GetAllGauges().Return(mockGaugeMap, nil)
			},
			wantGauges:   wantGaugeMap,
			wantCounters: wantCounterMap,
			wantErr:      nil,
		},
		{
			name: "failed to get counter map",
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetAllCounters().Return(nil, errTestGetAllCounters)
			},
			wantGauges:   nil,
			wantCounters: nil,
			wantErr:      errTestGetAllCounters,
		},
		{
			name: "got counter map, but failed to get gauge map",
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetAllCounters().Return(mockCounterMap, nil)
				repo.EXPECT().GetAllGauges().Return(nil, errTestGetAllGauges)
			},
			wantGauges:   nil,
			wantCounters: nil,
			wantErr:      errTestGetAllGauges,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockRepository(ctrl)

			svc := New(repo, dbPinger, zap.NewNop())
			tt.setup(repo)

			gotGauges, gotCounters, err := svc.ListAllMetrics()

			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("Service.ListAllMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("Service.ListAllMetrics() unexpected error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(gotGauges, tt.wantGauges) {
				t.Errorf("Service.ListAllMetrics() got = %v, want %v", gotGauges, tt.wantGauges)
			}
			if !reflect.DeepEqual(gotCounters, tt.wantCounters) {
				t.Errorf("Service.ListAllMetrics() gotCounters = %v, want %v", gotCounters, tt.wantCounters)
			}
		})
	}
}

func TestService_GetCounter(t *testing.T) {
	tests := []struct {
		name             string
		metricName       string
		setup            func(repo *mocks.MockRepository)
		wantCounterValue int64
		wantErr          error
	}{
		{
			name:       "returns counter value",
			metricName: testCounterMetricName,
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetCounter(testCounterMetricName).Return(mockCounterValue, nil)
			},
			wantCounterValue: wantCounterValue,
			wantErr:          nil,
		},
		{
			name:       "metric not found",
			metricName: testCounterMetricName,
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetCounter(testCounterMetricName).Return(int64(0), model.ErrMetricNotFound)
			},
			wantCounterValue: 0,
			wantErr:          model.ErrMetricNotFound,
		},
		{
			name:       "repository error is propagated",
			metricName: testCounterMetricName,
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetCounter(testCounterMetricName).Return(int64(0), errTestGetCounter)
			},
			wantCounterValue: 0,
			wantErr:          errTestGetCounter,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockRepository(ctrl)

			svc := New(repo, dbPinger, zap.NewNop())
			tt.setup(repo)

			got, err := svc.GetCounter(tt.metricName)
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("Service.GetCounter() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("Service.GetCounter() unexpected error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.wantCounterValue {
				t.Errorf("Service.GetCounter() = %v, want %v", got, tt.wantCounterValue)
			}
		})
	}
}

func TestService_GetGauge(t *testing.T) {
	tests := []struct {
		name           string
		metricName     string
		setup          func(repo *mocks.MockRepository)
		wantGaugeValue float64
		wantErr        error
	}{
		{
			name:       "returns gauge value",
			metricName: testGaugeMetricName,
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetGauge(testGaugeMetricName).Return(mockGaugeValue, nil)
			},
			wantGaugeValue: wantGaugeValue,
			wantErr:        nil,
		},
		{
			name:       "metric not found",
			metricName: testGaugeMetricName,
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetGauge(testGaugeMetricName).Return(float64(0), model.ErrMetricNotFound)
			},
			wantGaugeValue: 0,
			wantErr:        model.ErrMetricNotFound,
		},
		{
			name:       "repository error is propagated",
			metricName: testGaugeMetricName,
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetGauge(testGaugeMetricName).Return(float64(0), errTestGetGauge)
			},
			wantGaugeValue: 0,
			wantErr:        errTestGetGauge,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockRepository(ctrl)

			svc := New(repo, dbPinger, zap.NewNop())
			tt.setup(repo)

			got, err := svc.GetGauge(tt.metricName)
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("Service.GetGauge() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("Service.GetGauge() unexpected error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.wantGaugeValue {
				t.Errorf("Service.GetGauge() = %v, want %v", got, tt.wantGaugeValue)
			}
		})
	}
}

func TestService_SetGauge(t *testing.T) {
	tests := []struct {
		name       string
		metricName string
		value      float64
		setup      func(repo *mocks.MockRepository)
		wantErr    error
	}{
		{
			name:       "sets gauge value",
			metricName: testGaugeMetricName,
			value:      mockGaugeValue,
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().SetGauge(testGaugeMetricName, mockGaugeValue).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:       "metric not found",
			metricName: testGaugeMetricName,
			value:      mockGaugeValue,
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().SetGauge(testGaugeMetricName, mockGaugeValue).Return(model.ErrMetricNotFound)
			},
			wantErr: model.ErrMetricNotFound,
		},
		{
			name:       "repository error is propagated",
			metricName: testGaugeMetricName,
			value:      mockGaugeValue,
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().SetGauge(testGaugeMetricName, mockGaugeValue).Return(errTestSetGauge)
			},
			wantErr: errTestSetGauge,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockRepository(ctrl)

			svc := New(repo, dbPinger, zap.NewNop())
			tt.setup(repo)

			err := svc.SetGauge(tt.metricName, tt.value)
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("Service.SetGauge() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("Service.SetGauge() unexpected error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_AddCounter(t *testing.T) {
	tests := []struct {
		name             string
		metricName       string
		delta            int64
		setup            func(repo *mocks.MockRepository)
		wantCounterValue int64
		wantErr          error
	}{
		{
			name:       "returns new counter value",
			metricName: testCounterMetricName,
			delta:      mockCounterValue,
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().AddCounter(testCounterMetricName, mockCounterValue).Return(mockCounterValue, nil)
			},
			wantCounterValue: wantCounterValue,
			wantErr:          nil,
		},
		{
			name:       "metric not found",
			metricName: testCounterMetricName,
			delta:      mockCounterValue,
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().AddCounter(testCounterMetricName, mockCounterValue).Return(int64(0), model.ErrMetricNotFound)
			},
			wantCounterValue: 0,
			wantErr:          model.ErrMetricNotFound,
		},
		{
			name:       "repository error is propagated",
			metricName: testCounterMetricName,
			delta:      mockCounterValue,
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().AddCounter(testCounterMetricName, mockCounterValue).Return(int64(0), errTestAddCounter)
			},
			wantCounterValue: 0,
			wantErr:          errTestAddCounter,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockRepository(ctrl)

			svc := New(repo, dbPinger, zap.NewNop())
			tt.setup(repo)

			got, err := svc.AddCounter(tt.metricName, tt.delta)
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("Service.AddCounter() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("Service.AddCounter() unexpected error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.wantCounterValue {
				t.Errorf("Service.AddCounter() = %v, want %v", got, tt.wantCounterValue)
			}
		})
	}
}

func TestService_UpdateMetricsBatch(t *testing.T) {
	tests := []struct {
		name    string
		metrics []model.Metric
		setup   func(repo *mocks.MockRepository)
		wantErr error
	}{
		{
			name:    "updates metrics",
			metrics: mockMetricsBatch,
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().UpdateMetrics(gomock.Any(), mockMetricsBatch).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:    "repository error is propagated",
			metrics: mockMetricsBatch,
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().UpdateMetrics(gomock.Any(), mockMetricsBatch).Return(errTestUpdateMetrics)
			},
			wantErr: errTestUpdateMetrics,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockRepository(ctrl)

			svc := New(repo, dbPinger, zap.NewNop())
			tt.setup(repo)

			err := svc.UpdateMetricsBatch(t.Context(), tt.metrics)
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("Service.UpdateMetricsBatch() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("Service.UpdateMetricsBatch() unexpected error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_UpdateMetricsBatchCancelled(t *testing.T) {
	tests := []struct {
		name    string
		metrics []model.Metric
		setup   func(repo *mocks.MockRepository)
		wantErr error
	}{
		{
			name:    "context cancelled",
			metrics: mockMetricsBatch,
			setup: func(repo *mocks.MockRepository) {
				repo.EXPECT().UpdateMetrics(gomock.Any(), mockMetricsBatch).DoAndReturn(func(ctx context.Context, _ []model.Metric) error {
					<-ctx.Done()
					return ctx.Err()
				})
			},
			wantErr: context.Canceled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockRepository(ctrl)

			svc := New(repo, dbPinger, zap.NewNop())
			tt.setup(repo)

			ctx, cancel := context.WithCancel(t.Context())
			cancel()

			err := svc.UpdateMetricsBatch(ctx, tt.metrics)
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("Service.UpdateMetricsBatch() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("Service.UpdateMetricsBatch() unexpected error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
