package agent

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestSetGauge verifies that SetGauge stores the value and overwrites on subsequent calls.
func TestSetGauge(t *testing.T) {
	ms := NewAgentStorage()
	for _, m := range GaugeMetrics {
		t.Run(m, func(t *testing.T) {
			firstRandomValue := rand.Float64()
			ms.SetGauge(m, firstRandomValue)
			firstGet, firstErr := ms.GetGauge(m)
			require.NoError(t, firstErr)
			require.Equal(t, firstRandomValue, firstGet)

			secondRandomValue := rand.Float64()
			ms.SetGauge(m, secondRandomValue)
			secondGet, secondErr := ms.GetGauge(m)
			require.NoError(t, secondErr)
			require.Equal(t, secondRandomValue, secondGet)
		})
	}
}

// TestAddCounter verifies that AddCounter increments the value cumulatively.
func TestAddCounter(t *testing.T) {
	ms := NewAgentStorage()
	for _, m := range CounterMetrics {
		t.Run(m, func(t *testing.T) {
			firstRandomValue := rand.Int64N(10)
			ms.AddCounter(m, firstRandomValue)
			firstGet, firstErr := ms.GetCounter(m)
			require.NoError(t, firstErr)
			require.Equal(t, firstRandomValue, firstGet)

			secondRandomValue := rand.Int64N(10)
			ms.AddCounter(m, secondRandomValue)
			secondGet, secondErr := ms.GetCounter(m)
			require.NoError(t, secondErr)
			require.Equal(t, firstRandomValue+secondRandomValue, secondGet)
		})
	}
}

// TestResetCounter verifies that ResetCounter sets the counter value to zero.
func TestResetCounter(t *testing.T) {
	ms := NewAgentStorage()
	for _, m := range CounterMetrics {
		t.Run(m, func(t *testing.T) {
			// Set random value for m and compare
			randomValue := rand.Int64N(100)
			ms.AddCounter(m, randomValue)
			got, err := ms.GetCounter(m)
			require.NoError(t, err)
			require.Equal(t, randomValue, got)

			// Set zero for m and compare
			ms.ResetCounter(m)
			got, err = ms.GetCounter(m)
			require.NoError(t, err)
			require.Equal(t, int64(0), got)
		})
	}
}

// TestGetGaugeError verifies that GetGauge returns an error for a nonexistent key.
func TestGetGaugeError(t *testing.T) {
	ms := NewAgentStorage()
	t.Run("nonexistent gauge", func(t *testing.T) {
		_, err := ms.GetGauge("__nonexistent__")
		require.EqualError(t, err, fmt.Sprintf("Gauge %s not found", "__nonexistent__"))
	})
}

// TestGetCounterError verifies that GetCounter returns an error for a nonexistent key.
func TestGetCounterError(t *testing.T) {
	ms := NewAgentStorage()
	t.Run("nonexistent counter", func(t *testing.T) {
		_, err := ms.GetCounter("__nonexistent__")
		require.EqualError(t, err, fmt.Sprintf("Counter %s not found", "__nonexistent__"))
	})
}

// TestGetAllGaugesReturnsCopy verifies that GetAllGauges returns a copy, not a reference.
func TestGetAllGaugesReturnsCopy(t *testing.T) {
	ms := NewAgentStorage()
	gaugeCopy := ms.GetAllGauges()

	for _, m := range GaugeMetrics {
		t.Run(m, func(t *testing.T) {
			firstRandomValue := rand.Float64()
			ms.SetGauge(m, firstRandomValue)
			require.NotEqual(t, gaugeCopy[m], firstRandomValue)

			secondRandomValue := rand.Float64()
			ms.SetGauge(m, secondRandomValue)
			require.NotEqual(t, gaugeCopy[m], secondRandomValue)
		})
	}
}

// TestGetAllCountersReturnsCopy verifies that GetAllCounters returns a copy, not a reference.
func TestGetAllCountersReturnsCopy(t *testing.T) {
	ms := NewAgentStorage()
	counterCopy := ms.GetAllCounters()

	for _, m := range CounterMetrics {
		t.Run(m, func(t *testing.T) {
			firstRandomValue := rand.Int64()
			ms.AddCounter(m, firstRandomValue)
			require.NotEqual(t, counterCopy[m], firstRandomValue)

			secondRandomValue := rand.Int64()
			ms.AddCounter(m, secondRandomValue)
			require.NotEqual(t, counterCopy[m], secondRandomValue)
		})
	}
}

// TestGetAllGauges verifies that GetAllGauges returns a map containing all stored values.
func TestGetAllGauges(t *testing.T) {
	ms := NewAgentStorage()
	for _, m := range GaugeMetrics {
		t.Run(m, func(t *testing.T) {
			rv := rand.Float64()
			ms.SetGauge(m, rv)
			got, err := ms.GetGauge(m)

			gauges := ms.GetAllGauges()
			require.NoError(t, err)
			require.Equal(t, gauges[m], got)
		})

	}
}

// TestGetAllCounters verifies that GetAllCounters returns a map containing all stored values.
func TestGetAllCounters(t *testing.T) {
	ms := NewAgentStorage()
	for _, m := range CounterMetrics {
		t.Run(m, func(t *testing.T) {
			rv := rand.Int64()
			ms.AddCounter(m, rv)
			got, err := ms.GetCounter(m)

			counters := ms.GetAllCounters()
			require.NoError(t, err)
			require.Equal(t, counters[m], got)
		})
	}
}
