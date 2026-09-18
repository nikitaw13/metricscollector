package agent

import (
	"testing"
)

// TestAllMetricsExist verifies that Collect returns all required gauge and counter metrics in the batch.
func TestAllMetricsExist(t *testing.T) {
	metrics := Collect()

	found := make(map[string]bool)
	for _, m := range metrics {
		found[m.ID] = true
	}

	for _, name := range GaugeMetrics {
		if !found[name] {
			t.Errorf("gauge metric %q not found", name)
		}
	}
	for _, name := range CounterMetrics {
		if !found[name] {
			t.Errorf("counter metric %q not found", name)
		}
	}
}

// TestPollCountPerSnapshot verifies that each collected snapshot carries PollCount with a delta of 1.
func TestPollCountPerSnapshot(t *testing.T) {
	metrics := Collect()

	for _, m := range metrics {
		if m.ID == "PollCount" {
			if m.Delta == nil {
				t.Fatal("PollCount delta is nil")
			}
			if *m.Delta != 1 {
				t.Errorf("PollCount delta = %d, want 1", *m.Delta)
			}
			return
		}
	}
	t.Error("PollCount not found in batch")
}
