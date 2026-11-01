package agent

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/nikitaw13/metricscollector/internal/model"
)

// TestRuntimeMetricsExist verifies that CollectRuntimeMetrics returns all required gauge and counter metrics in the batch.
func TestRuntimeMetricsExist(t *testing.T) {
	metrics := CollectRuntimeMetrics()

	found := make(map[string]bool)
	for _, metric := range metrics {
		found[metric.ID] = true
	}

	for _, name := range gaugeMetricNames {
		if !found[name] {
			t.Errorf("gauge metric %q not found", name)
		}
	}
	for _, name := range counterMetricNames {
		if !found[name] {
			t.Errorf("counter metric %q not found", name)
		}
	}
}

// TestPollCountPerSnapshot verifies that each collected snapshot carries PollCount with a delta of 1.
func TestPollCountPerSnapshot(t *testing.T) {
	metrics := CollectRuntimeMetrics()

	for _, metric := range metrics {
		if metric.ID == "PollCount" {
			if metric.Delta == nil {
				t.Fatal("PollCount delta is nil")
			}
			if *metric.Delta != 1 {
				t.Errorf("PollCount delta = %d, want 1", *metric.Delta)
			}
			return
		}
	}
	t.Error("PollCount not found in batch")
}

// TestSystemMetricsExist verifies that the batch returned by
// CollectSystemMetrics contains every gopsutil gauge metric, including one
// CPUutilization metric per CPU core.
func TestSystemMetricsExist(t *testing.T) {
	metrics := CollectSystemMetrics()

	found := make(map[string]bool)
	for _, metric := range metrics {
		if metric.Type != model.Gauge {
			t.Errorf("metric %q has type %q, want gauge", metric.ID, metric.Type)
		}
		if metric.Value == nil {
			t.Errorf("gauge metric %q has nil value", metric.ID)
		}
		found[metric.ID] = true
	}

	expected := []string{"TotalMemory", "FreeMemory"}
	for i := 1; i <= runtime.NumCPU(); i++ {
		expected = append(expected, fmt.Sprintf("CPUutilization%d", i))
	}

	for _, name := range expected {
		if !found[name] {
			t.Errorf("gauge metric %q not found in batch", name)
		}
	}
}
