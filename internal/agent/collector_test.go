package agent

import "testing"

func TestGet_Run(t *testing.T) {
	storage := NewAgentStorage()
	c := &Collector{Storage: storage}
	c.Run()

	// Проверяй напрямую — ключ существует и значение разумное
	val, _ := storage.GetGauge("Alloc")
	if val == 0 {
		t.Error("Alloc should be > 0")
	}

	// PollCount должен быть 1
	pollCount, _ := storage.GetCounter("PollCount")
	if pollCount != 1 {
		t.Errorf("PollCount = %d, want 1", pollCount)
	}

	c.Run()

	// PollCount должен быть 2
	pollCount, _ = storage.GetCounter("PollCount")
	if pollCount != 2 {
		t.Errorf("PollCount = %d, want 2", pollCount)
	}
}
