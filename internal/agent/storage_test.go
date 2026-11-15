package agent

import (
	"math/rand/v2"
	"testing"

	"github.com/nikitaw13/metricscollector/internal/model"
	"github.com/stretchr/testify/require"
)

// TestSetGauge verifies that SetGauge stores the value and overwrites on subsequent calls.
func TestSetGauge(t *testing.T) {
	storage := NewAgentStorage()
	for _, metricName := range gaugeMetricNames {
		t.Run(metricName, func(t *testing.T) {
			t.Parallel()
			firstRandomValue := rand.Float64()
			storage.SetGauge(metricName, firstRandomValue)
			firstGet, firstErr := storage.GetGauge(metricName)
			require.NoError(t, firstErr)
			require.Equal(t, firstRandomValue, firstGet)

			secondRandomValue := rand.Float64()
			storage.SetGauge(metricName, secondRandomValue)
			secondGet, secondErr := storage.GetGauge(metricName)
			require.NoError(t, secondErr)
			require.Equal(t, secondRandomValue, secondGet)
		})
	}
}

// TestAddCounter verifies that AddCounter increments the value cumulatively.
func TestAddCounter(t *testing.T) {
	storage := NewAgentStorage()
	for _, metricName := range counterMetricNames {
		t.Run(metricName, func(t *testing.T) {
			t.Parallel()
			firstRandomValue := rand.Int64N(10)
			storage.AddCounter(metricName, firstRandomValue)
			firstGet, firstErr := storage.GetCounter(metricName)
			require.NoError(t, firstErr)
			require.Equal(t, firstRandomValue, firstGet)

			secondRandomValue := rand.Int64N(10)
			storage.AddCounter(metricName, secondRandomValue)
			secondGet, secondErr := storage.GetCounter(metricName)
			require.NoError(t, secondErr)
			require.Equal(t, firstRandomValue+secondRandomValue, secondGet)
		})
	}
}

// TestGetGaugeError verifies that GetGauge returns an error for a nonexistent key.
func TestGetGaugeError(t *testing.T) {
	storage := NewAgentStorage()
	t.Run("nonexistent gauge", func(t *testing.T) {
		_, err := storage.GetGauge("__nonexistent__")
		require.ErrorIs(t, err, model.ErrMetricNotFound)
	})
}

// TestGetCounterError verifies that GetCounter returns an error for a nonexistent key.
func TestGetCounterError(t *testing.T) {
	storage := NewAgentStorage()
	t.Run("nonexistent counter", func(t *testing.T) {
		_, err := storage.GetCounter("__nonexistent__")
		require.ErrorIs(t, err, model.ErrMetricNotFound)
	})
}

// TestGetAllGaugesReturnsCopy verifies that GetAllGauges returns a copy, not a reference.
func TestGetAllGaugesReturnsCopy(t *testing.T) {
	storage := NewAgentStorage()
	gaugeCopy := storage.GetAllGauges()

	for _, metricName := range gaugeMetricNames {
		t.Run(metricName, func(t *testing.T) {
			t.Parallel()
			firstRandomValue := rand.Float64()
			storage.SetGauge(metricName, firstRandomValue)
			require.NotEqual(t, gaugeCopy[metricName], firstRandomValue)

			secondRandomValue := rand.Float64()
			storage.SetGauge(metricName, secondRandomValue)
			require.NotEqual(t, gaugeCopy[metricName], secondRandomValue)
		})
	}
}

// TestGetAllCountersReturnsCopy verifies that GetAllCounters returns a copy, not a reference.
func TestGetAllCountersReturnsCopy(t *testing.T) {
	storage := NewAgentStorage()
	counterCopy := storage.GetAllCounters()

	for _, metricName := range counterMetricNames {
		t.Run(metricName, func(t *testing.T) {
			t.Parallel()
			firstRandomValue := rand.Int64()
			storage.AddCounter(metricName, firstRandomValue)
			require.NotEqual(t, counterCopy[metricName], firstRandomValue)

			secondRandomValue := rand.Int64()
			storage.AddCounter(metricName, secondRandomValue)
			require.NotEqual(t, counterCopy[metricName], secondRandomValue)
		})
	}
}

// TestGetAllGauges verifies that GetAllGauges returns a map containing all stored values.
func TestGetAllGauges(t *testing.T) {
	storage := NewAgentStorage()
	for _, metricName := range gaugeMetricNames {
		t.Run(metricName, func(t *testing.T) {
			t.Parallel()
			initialValue := rand.Float64()
			storage.SetGauge(metricName, initialValue)
			got, err := storage.GetGauge(metricName)

			gauges := storage.GetAllGauges()
			require.NoError(t, err)
			require.Equal(t, gauges[metricName], got)
		})

	}
}

// TestGetAllCounters verifies that GetAllCounters returns a map containing all stored values.
func TestGetAllCounters(t *testing.T) {
	storage := NewAgentStorage()
	for _, metricName := range counterMetricNames {
		t.Run(metricName, func(t *testing.T) {
			t.Parallel()
			initialValue := rand.Int64()
			storage.AddCounter(metricName, initialValue)
			got, err := storage.GetCounter(metricName)

			counters := storage.GetAllCounters()
			require.NoError(t, err)
			require.Equal(t, counters[metricName], got)
		})
	}
}
