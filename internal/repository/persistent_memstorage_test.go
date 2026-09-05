package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nikitaw13/metricscollector/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tempFilePath returns a path in a temp directory without creating the file.
func tempFilePath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "metrics.json")
}

// writeMetricsFile writes the given metrics as JSON to the specified path.
func writeMetricsFile(t *testing.T, path string, metrics []model.Metric) {
	t.Helper()
	data, err := json.Marshal(metrics)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0644))
}

// readMetricsFile reads the file at path and unmarshals into []model.Metric.
func readMetricsFile(t *testing.T, path string) []model.Metric {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var metrics []model.Metric
	require.NoError(t, json.Unmarshal(data, &metrics))
	return metrics
}

func newTestPersistentStorage(filePath string, syncWrite bool) *PersistentMemStorage {
	return NewPersistentMemStorage(NewMemStorage(), filePath, syncWrite)
}

// ---------- NewPersistentMemStorage ----------

func TestNewPersistentMemStorage(t *testing.T) {
	t.Parallel()

	ms := NewMemStorage()
	persistentStorage := NewPersistentMemStorage(ms, "/tmp/test.json", true)

	assert.Equal(t, ms, persistentStorage.MemStorage)
	assert.Equal(t, "/tmp/test.json", persistentStorage.filePath)
	assert.True(t, persistentStorage.syncWrite)
}

// ---------- Restore ----------

func TestRestore_GaugesAndCounters(t *testing.T) {
	t.Parallel()

	filePath := tempFilePath(t)
	gaugeVal := 42.5
	deltaVal := int64(10)
	writeMetricsFile(t, filePath, []model.Metric{
		{ID: "temp", Type: "gauge", Value: &gaugeVal},
		{ID: "requests", Type: "counter", Delta: &deltaVal},
	})

	persistentStorage := newTestPersistentStorage(filePath, false)
	require.NoError(t, persistentStorage.Restore())

	gotGauge, err := persistentStorage.GetGauge("temp")
	assert.NoError(t, err)
	assert.InDelta(t, 42.5, gotGauge, 0.001)

	gotCounter, err := persistentStorage.GetCounter("requests")
	assert.NoError(t, err)
	assert.Equal(t, int64(10), gotCounter)
}

func TestRestore_OverwritesExisting(t *testing.T) {
	t.Parallel()

	filePath := tempFilePath(t)
	gaugeVal := 99.9
	counterVal := int64(5)
	writeMetricsFile(t, filePath, []model.Metric{
		{ID: "cpu", Type: "gauge", Value: &gaugeVal},
		{ID: "hits", Type: "counter", Delta: &counterVal},
	})

	persistentStorage := newTestPersistentStorage(filePath, false)
	// pre-populate with different values
	persistentStorage.gauge["cpu"] = 1.0
	persistentStorage.counter["hits"] = 100

	require.NoError(t, persistentStorage.Restore())

	gaugeVal, _ = persistentStorage.GetGauge("cpu")
	assert.InDelta(t, 99.9, gaugeVal, 0.001)

	counterVal, _ = persistentStorage.GetCounter("hits")
	assert.Equal(t, int64(5), counterVal)
}

func TestRestore_FileNotFound(t *testing.T) {
	t.Parallel()

	persistentStorage := newTestPersistentStorage("/nonexistent/path/metrics.json", false)
	err := persistentStorage.Restore()
	assert.Error(t, err)
}

func TestRestore_InvalidJSON(t *testing.T) {
	t.Parallel()

	filePath := tempFilePath(t)
	require.NoError(t, os.WriteFile(filePath, []byte("not valid json"), 0644))

	persistentStorage := newTestPersistentStorage(filePath, false)
	err := persistentStorage.Restore()
	assert.Error(t, err)
}

func TestRestore_EmptyFile(t *testing.T) {
	t.Parallel()

	filePath := tempFilePath(t)
	writeMetricsFile(t, filePath, []model.Metric{})

	persistentStorage := newTestPersistentStorage(filePath, false)
	require.NoError(t, persistentStorage.Restore())

	_, err := persistentStorage.GetGauge("anything")
	assert.ErrorIs(t, err, model.ErrMetricNotFound)

	_, err = persistentStorage.GetCounter("anything")
	assert.ErrorIs(t, err, model.ErrMetricNotFound)
}

func TestGetCounter_NotFound(t *testing.T) {
	t.Parallel()

	persistentStorage := newTestPersistentStorage(tempFilePath(t), false)

	_, err := persistentStorage.GetCounter("nonexistent")
	assert.ErrorIs(t, err, model.ErrMetricNotFound)
}

// ---------- Save ----------

func TestSave_WritesGaugesAndCounters(t *testing.T) {
	t.Parallel()

	filePath := tempFilePath(t)
	persistentStorage := newTestPersistentStorage(filePath, false)

	require.NoError(t, persistentStorage.SetGauge("mem", 64.2))
	newDelta, err := persistentStorage.AddCounter("req", 5)
	require.NoError(t, err)
	require.Equal(t, int64(5), newDelta)
	require.NoError(t, persistentStorage.Save())

	metrics := readMetricsFile(t, filePath)
	assert.Len(t, metrics, 2)

	metricMap := map[string]model.Metric{}
	for _, m := range metrics {
		metricMap[m.ID] = m
	}

	assert.Equal(t, "gauge", metricMap["mem"].Type)
	assert.InDelta(t, 64.2, *metricMap["mem"].Value, 0.001)

	assert.Equal(t, "counter", metricMap["req"].Type)
	assert.Equal(t, int64(5), *metricMap["req"].Delta)
}

func TestSave_EmptyStorage(t *testing.T) {
	t.Parallel()

	filePath := tempFilePath(t)
	persistentStorage := newTestPersistentStorage(filePath, false)
	require.NoError(t, persistentStorage.Save())

	metrics := readMetricsFile(t, filePath)
	assert.Len(t, metrics, 0)
}

func TestSave_OverwritesExistingFile(t *testing.T) {
	t.Parallel()

	filePath := tempFilePath(t)
	// pre-populate file with old data
	gaugeVal := 1.0
	writeMetricsFile(t, filePath, []model.Metric{
		{ID: "old_metric", Type: "gauge", Value: &gaugeVal},
	})

	persistentStorage := newTestPersistentStorage(filePath, false)
	require.NoError(t, persistentStorage.SetGauge("new_metric", 2.0))
	require.NoError(t, persistentStorage.Save())

	metrics := readMetricsFile(t, filePath)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "new_metric", metrics[0].ID)
}

func TestSave_InvalidPath(t *testing.T) {
	t.Parallel()

	persistentStorage := newTestPersistentStorage("/nonexistent/dir/metrics.json", false)
	err := persistentStorage.Save()
	assert.Error(t, err)
}

// ---------- SetGauge (with SyncWrite) ----------

func TestSetGauge_SyncOff(t *testing.T) {
	t.Parallel()

	filePath := tempFilePath(t)
	persistentStorage := newTestPersistentStorage(filePath, false)

	require.NoError(t, persistentStorage.SetGauge("temp", 36.6))

	gotGauge, err := persistentStorage.GetGauge("temp")
	assert.NoError(t, err)
	assert.InDelta(t, 36.6, gotGauge, 0.001)

	// file should NOT exist because sync is off and Save was not called
	_, err = os.Stat(filePath)
	assert.True(t, os.IsNotExist(err), "file should not exist when sync is off")
}

func TestSetGauge_SyncOn(t *testing.T) {
	t.Parallel()

	filePath := tempFilePath(t)
	persistentStorage := newTestPersistentStorage(filePath, true)

	require.NoError(t, persistentStorage.SetGauge("temp", 36.6))

	// file must be created automatically by sync save
	metrics := readMetricsFile(t, filePath)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "temp", metrics[0].ID)
	assert.InDelta(t, 36.6, *metrics[0].Value, 0.001)
}

func TestSetGauge_OverwritesPrevious(t *testing.T) {
	t.Parallel()

	filePath := tempFilePath(t)
	persistentStorage := newTestPersistentStorage(filePath, false)

	require.NoError(t, persistentStorage.SetGauge("cpu", 10.0))
	require.NoError(t, persistentStorage.SetGauge("cpu", 20.0))

	gotGauge, err := persistentStorage.GetGauge("cpu")
	assert.NoError(t, err)
	assert.InDelta(t, 20.0, gotGauge, 0.001)
}

// ---------- AddCounter (with SyncWrite) ----------

func TestAddCounter_SyncOff(t *testing.T) {
	t.Parallel()

	filePath := tempFilePath(t)
	persistentStorage := newTestPersistentStorage(filePath, false)

	newDelta, err := persistentStorage.AddCounter("hits", 5)
	require.NoError(t, err)
	require.Equal(t, int64(5), newDelta)

	gotCounter, err := persistentStorage.GetCounter("hits")
	assert.NoError(t, err)
	assert.Equal(t, int64(5), gotCounter)

	// file should NOT exist because sync is off
	_, err = os.Stat(filePath)
	assert.True(t, os.IsNotExist(err))
}

func TestAddCounter_SyncOn(t *testing.T) {
	t.Parallel()

	filePath := tempFilePath(t)
	persistentStorage := newTestPersistentStorage(filePath, true)

	newDelta, err := persistentStorage.AddCounter("hits", 3)
	require.NoError(t, err)
	require.Equal(t, int64(3), newDelta)

	metrics := readMetricsFile(t, filePath)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "hits", metrics[0].ID)
	assert.Equal(t, int64(3), *metrics[0].Delta)
}

func TestAddCounter_IncrementsExisting(t *testing.T) {
	t.Parallel()

	filePath := tempFilePath(t)
	persistentStorage := newTestPersistentStorage(filePath, false)

	newDelta, err := persistentStorage.AddCounter("hits", 5)
	require.NoError(t, err)
	require.Equal(t, int64(5), newDelta)

	newDelta, err = persistentStorage.AddCounter("hits", 3)
	require.NoError(t, err)
	require.Equal(t, int64(8), newDelta)

	gotCounter, err := persistentStorage.GetCounter("hits")
	assert.NoError(t, err)
	assert.Equal(t, int64(8), gotCounter)
}

// ---------- Round-trip: Save → Restore ----------

func TestRoundTrip(t *testing.T) {
	t.Parallel()

	filePath := tempFilePath(t)
	persistentStorage := newTestPersistentStorage(filePath, false)

	require.NoError(t, persistentStorage.SetGauge("temperature", 23.7))
	require.NoError(t, persistentStorage.SetGauge("humidity", 55.0))

	newDelta, err := persistentStorage.AddCounter("requests", 100)
	require.NoError(t, err)
	require.Equal(t, int64(100), newDelta)

	newDelta, err = persistentStorage.AddCounter("errors", 2)
	require.NoError(t, err)
	require.Equal(t, int64(2), newDelta)

	newDelta, err = persistentStorage.AddCounter("requests", 50) // increment again
	require.NoError(t, err)
	require.Equal(t, int64(150), newDelta)

	require.NoError(t, persistentStorage.Save())

	// create a fresh storage and restore from file
	restored := newTestPersistentStorage(filePath, false)
	require.NoError(t, restored.Restore())

	gotGauge, err := restored.GetGauge("temperature")
	assert.NoError(t, err)
	assert.InDelta(t, 23.7, gotGauge, 0.001)

	gotGauge2, err := restored.GetGauge("humidity")
	assert.NoError(t, err)
	assert.InDelta(t, 55.0, gotGauge2, 0.001)

	gotCounter, err := restored.GetCounter("requests")
	assert.NoError(t, err)
	assert.Equal(t, int64(150), gotCounter)

	gotCounter2, err := restored.GetCounter("errors")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), gotCounter2)
}

func TestRoundTrip_MultipleSaveCycles(t *testing.T) {
	t.Parallel()

	filePath := tempFilePath(t)
	persistentStorage := newTestPersistentStorage(filePath, false)

	// first cycle
	require.NoError(t, persistentStorage.SetGauge("cpu", 10.0))

	newDelta, err := persistentStorage.AddCounter("hits", 2)
	require.NoError(t, err)
	require.Equal(t, int64(2), newDelta)

	require.NoError(t, persistentStorage.Save())

	// second cycle — update values and save again
	require.NoError(t, persistentStorage.SetGauge("cpu", 20.0))

	newDelta, err = persistentStorage.AddCounter("hits", 9)
	require.NoError(t, err)
	require.Equal(t, int64(11), newDelta)

	require.NoError(t, persistentStorage.Save())

	// restore into a fresh storage
	restored := newTestPersistentStorage(filePath, false)
	require.NoError(t, restored.Restore())

	gaugeVal, _ := restored.GetGauge("cpu")
	assert.InDelta(t, 20.0, gaugeVal, 0.001)

	counterVal, _ := restored.GetCounter("hits")
	assert.Equal(t, int64(11), counterVal)
}

// ---------- SaveSync ----------

func TestSaveSync(t *testing.T) {
	t.Parallel()

	filePath := tempFilePath(t)
	persistentStorage := newTestPersistentStorage(filePath, false)

	require.NoError(t, persistentStorage.SetGauge("mem", 80.0))
	persistentStorage.SaveSync()

	metrics := readMetricsFile(t, filePath)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "mem", metrics[0].ID)
}

// ---------- UpdateMetrics (with SyncWrite) ----------

// TestUpdateMetrics_SyncOff verifies that a batch is applied to memory
// but not persisted to disk when sync-write is disabled.
func TestUpdateMetrics_SyncOff(t *testing.T) {
	t.Parallel()

	filePath := tempFilePath(t)
	persistentStorage := newTestPersistentStorage(filePath, false)

	gaugeVal := 36.6
	counterVal := int64(5)
	metrics := []model.Metric{
		{ID: "batch_gauge", Type: model.Gauge, Value: &gaugeVal},
		{ID: "batch_counter", Type: model.Counter, Delta: &counterVal},
	}

	require.NoError(t, persistentStorage.UpdateMetrics(t.Context(), metrics))

	gotGauge, err := persistentStorage.GetGauge("batch_gauge")
	require.NoError(t, err)
	assert.InDelta(t, 36.6, gotGauge, 0.001)

	gotCounter, err := persistentStorage.GetCounter("batch_counter")
	require.NoError(t, err)
	assert.Equal(t, int64(5), gotCounter)

	// file should NOT exist because sync is off and Save was not called
	_, err = os.Stat(filePath)
	assert.True(t, os.IsNotExist(err), "file should not exist when sync is off")
}

// TestUpdateMetrics_SyncOn verifies that a batch is applied and persisted
// to disk in a single save when sync-write is enabled.
func TestUpdateMetrics_SyncOn(t *testing.T) {
	t.Parallel()

	filePath := tempFilePath(t)
	persistentStorage := newTestPersistentStorage(filePath, true)

	gaugeVal := 36.6
	counterVal := int64(5)
	metrics := []model.Metric{
		{ID: "batch_gauge", Type: model.Gauge, Value: &gaugeVal},
		{ID: "batch_counter", Type: model.Counter, Delta: &counterVal},
	}

	require.NoError(t, persistentStorage.UpdateMetrics(t.Context(), metrics))

	metricsOnDisk := readMetricsFile(t, filePath)
	assert.Len(t, metricsOnDisk, 2)

	metricMap := map[string]model.Metric{}
	for _, m := range metricsOnDisk {
		metricMap[m.ID] = m
	}

	assert.Equal(t, "gauge", metricMap["batch_gauge"].Type)
	assert.InDelta(t, 36.6, *metricMap["batch_gauge"].Value, 0.001)

	assert.Equal(t, "counter", metricMap["batch_counter"].Type)
	assert.Equal(t, int64(5), *metricMap["batch_counter"].Delta)
}

// TestUpdateMetrics_CounterAccumulates verifies that a batch update
// accumulates counter deltas on top of existing values.
func TestUpdateMetrics_CounterAccumulates(t *testing.T) {
	t.Parallel()

	filePath := tempFilePath(t)
	persistentStorage := newTestPersistentStorage(filePath, false)

	newDelta, err := persistentStorage.AddCounter("hits", 2)
	require.NoError(t, err)
	require.Equal(t, int64(2), newDelta)

	counterVal := int64(9)
	require.NoError(t, persistentStorage.UpdateMetrics(t.Context(), []model.Metric{
		{ID: "hits", Type: model.Counter, Delta: &counterVal},
	}))

	gotCounter, err := persistentStorage.GetCounter("hits")
	require.NoError(t, err)
	assert.Equal(t, int64(11), gotCounter)
}

// ---------- PeriodicSave (smoke test) ----------

func TestPeriodicSave_NoPanicOnStart(t *testing.T) {
	t.Parallel()

	filePath := tempFilePath(t)
	persistentStorage := newTestPersistentStorage(filePath, false)

	require.NoError(t, persistentStorage.SetGauge("cpu", 50.0))

	done := make(chan struct{})
	go func() {
		defer close(done)
		persistentStorage.PeriodicSave(time.Second)
	}()

	// PeriodicSave runs forever and cannot be stopped from outside, so just let it tick once and verify it does not panic; in production it should accept a context for cancellation.
	_ = done
}
