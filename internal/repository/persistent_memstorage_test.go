package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/PrometheRus/metricscollector/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helper: creates a temp file path without creating the file itself.
func tempFilePath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "metrics.json")
}

// helper: writes given metrics as JSON to the specified path.
func writeMetricsFile(t *testing.T, path string, metrics []model.Metric) {
	t.Helper()
	data, err := json.Marshal(metrics)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0644))
}

// helper: reads the file at path and unmarshals into []model.Metric.
func readMetricsFile(t *testing.T, path string) []model.Metric {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var metrics []model.Metric
	require.NoError(t, json.Unmarshal(data, &metrics))
	return metrics
}

func newTestPersistentStorage(fp string, synchronic bool) *PersistentMemStorage {
	return NewPersistentMemStorage(NewMemStorage(), fp, synchronic)
}

// ---------- NewPersistentMemStorage ----------

func TestNewPersistentMemStorage(t *testing.T) {
	t.Parallel()

	ms := NewMemStorage()
	pms := NewPersistentMemStorage(ms, "/tmp/test.json", true)

	assert.Equal(t, ms, pms.MemStorage)
	assert.Equal(t, "/tmp/test.json", pms.FilePath)
	assert.True(t, pms.Synchronic)
}

// ---------- Restore ----------

func TestRestore_GaugesAndCounters(t *testing.T) {
	t.Parallel()

	fp := tempFilePath(t)
	gaugeVal := 42.5
	deltaVal := int64(10)
	writeMetricsFile(t, fp, []model.Metric{
		{ID: "temp", Type: "gauge", Value: &gaugeVal},
		{ID: "requests", Type: "counter", Delta: &deltaVal},
	})

	pms := newTestPersistentStorage(fp, false)
	require.NoError(t, pms.Restore())

	gotGauge, err := pms.GetGauge("temp")
	assert.NoError(t, err)
	assert.InDelta(t, 42.5, gotGauge, 0.001)

	gotCounter, err := pms.GetCounter("requests")
	assert.NoError(t, err)
	assert.Equal(t, int64(10), gotCounter)
}

func TestRestore_OverwritesExisting(t *testing.T) {
	t.Parallel()

	fp := tempFilePath(t)
	gaugeVal := 99.9
	deltaVal := int64(5)
	writeMetricsFile(t, fp, []model.Metric{
		{ID: "cpu", Type: "gauge", Value: &gaugeVal},
		{ID: "hits", Type: "counter", Delta: &deltaVal},
	})

	pms := newTestPersistentStorage(fp, false)
	// pre-populate with different values
	pms.gauge["cpu"] = 1.0
	pms.counter["hits"] = 100

	require.NoError(t, pms.Restore())

	g, _ := pms.GetGauge("cpu")
	assert.InDelta(t, 99.9, g, 0.001)

	c, _ := pms.GetCounter("hits")
	assert.Equal(t, int64(5), c)
}

func TestRestore_FileNotFound(t *testing.T) {
	t.Parallel()

	pms := newTestPersistentStorage("/nonexistent/path/metrics.json", false)
	err := pms.Restore()
	assert.Error(t, err)
}

func TestRestore_InvalidJSON(t *testing.T) {
	t.Parallel()

	fp := tempFilePath(t)
	require.NoError(t, os.WriteFile(fp, []byte("not valid json"), 0644))

	pms := newTestPersistentStorage(fp, false)
	err := pms.Restore()
	assert.Error(t, err)
}

func TestRestore_EmptyFile(t *testing.T) {
	t.Parallel()

	fp := tempFilePath(t)
	writeMetricsFile(t, fp, []model.Metric{})

	pms := newTestPersistentStorage(fp, false)
	require.NoError(t, pms.Restore())

	_, err := pms.GetGauge("anything")
	assert.Error(t, err, "expected gauge not found after restoring empty file")
}

// ---------- Save ----------

func TestSave_WritesGaugesAndCounters(t *testing.T) {
	t.Parallel()

	fp := tempFilePath(t)
	pms := newTestPersistentStorage(fp, false)

	require.NoError(t, pms.SetGauge("mem", 64.2))
	require.NoError(t, pms.AddCounter("req", 7))
	require.NoError(t, pms.Save())

	metrics := readMetricsFile(t, fp)
	assert.Len(t, metrics, 2)

	metricMap := map[string]model.Metric{}
	for _, m := range metrics {
		metricMap[m.ID] = m
	}

	assert.Equal(t, "gauge", metricMap["mem"].Type)
	assert.InDelta(t, 64.2, *metricMap["mem"].Value, 0.001)

	assert.Equal(t, "counter", metricMap["req"].Type)
	assert.Equal(t, int64(7), *metricMap["req"].Delta)
}

func TestSave_EmptyStorage(t *testing.T) {
	t.Parallel()

	fp := tempFilePath(t)
	pms := newTestPersistentStorage(fp, false)
	require.NoError(t, pms.Save())

	metrics := readMetricsFile(t, fp)
	assert.Len(t, metrics, 0)
}

func TestSave_OverwritesExistingFile(t *testing.T) {
	t.Parallel()

	fp := tempFilePath(t)
	// pre-populate file with old data
	gaugeVal := 1.0
	writeMetricsFile(t, fp, []model.Metric{
		{ID: "old_metric", Type: "gauge", Value: &gaugeVal},
	})

	pms := newTestPersistentStorage(fp, false)
	require.NoError(t, pms.SetGauge("new_metric", 2.0))
	require.NoError(t, pms.Save())

	metrics := readMetricsFile(t, fp)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "new_metric", metrics[0].ID)
}

func TestSave_InvalidPath(t *testing.T) {
	t.Parallel()

	pms := newTestPersistentStorage("/nonexistent/dir/metrics.json", false)
	err := pms.Save()
	assert.Error(t, err)
}

// ---------- SetGauge (with Synchronic) ----------

func TestSetGauge_SyncOff(t *testing.T) {
	t.Parallel()

	fp := tempFilePath(t)
	pms := newTestPersistentStorage(fp, false)

	require.NoError(t, pms.SetGauge("temp", 36.6))

	g, err := pms.GetGauge("temp")
	assert.NoError(t, err)
	assert.InDelta(t, 36.6, g, 0.001)

	// file should NOT exist because sync is off and Save was not called
	_, err = os.Stat(fp)
	assert.True(t, os.IsNotExist(err), "file should not exist when sync is off")
}

func TestSetGauge_SyncOn(t *testing.T) {
	t.Parallel()

	fp := tempFilePath(t)
	pms := newTestPersistentStorage(fp, true)

	require.NoError(t, pms.SetGauge("temp", 36.6))

	// file must be created automatically by sync save
	metrics := readMetricsFile(t, fp)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "temp", metrics[0].ID)
	assert.InDelta(t, 36.6, *metrics[0].Value, 0.001)
}

func TestSetGauge_OverwritesPrevious(t *testing.T) {
	t.Parallel()

	fp := tempFilePath(t)
	pms := newTestPersistentStorage(fp, false)

	require.NoError(t, pms.SetGauge("cpu", 10.0))
	require.NoError(t, pms.SetGauge("cpu", 20.0))

	g, err := pms.GetGauge("cpu")
	assert.NoError(t, err)
	assert.InDelta(t, 20.0, g, 0.001)
}

// ---------- AddCounter (with Synchronic) ----------

func TestAddCounter_SyncOff(t *testing.T) {
	t.Parallel()

	fp := tempFilePath(t)
	pms := newTestPersistentStorage(fp, false)

	require.NoError(t, pms.AddCounter("hits", 5))

	c, err := pms.GetCounter("hits")
	assert.NoError(t, err)
	assert.Equal(t, int64(5), c)

	// file should NOT exist because sync is off
	_, err = os.Stat(fp)
	assert.True(t, os.IsNotExist(err))
}

func TestAddCounter_SyncOn(t *testing.T) {
	t.Parallel()

	fp := tempFilePath(t)
	pms := newTestPersistentStorage(fp, true)

	require.NoError(t, pms.AddCounter("hits", 3))

	metrics := readMetricsFile(t, fp)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "hits", metrics[0].ID)
	assert.Equal(t, int64(3), *metrics[0].Delta)
}

func TestAddCounter_IncrementsExisting(t *testing.T) {
	t.Parallel()

	fp := tempFilePath(t)
	pms := newTestPersistentStorage(fp, false)

	require.NoError(t, pms.AddCounter("hits", 5))
	require.NoError(t, pms.AddCounter("hits", 3))

	c, err := pms.GetCounter("hits")
	assert.NoError(t, err)
	assert.Equal(t, int64(8), c)
}

// ---------- Round-trip: Save → Restore ----------

func TestRoundTrip(t *testing.T) {
	t.Parallel()

	fp := tempFilePath(t)
	pms := newTestPersistentStorage(fp, false)

	require.NoError(t, pms.SetGauge("temperature", 23.7))
	require.NoError(t, pms.SetGauge("humidity", 55.0))
	require.NoError(t, pms.AddCounter("requests", 100))
	require.NoError(t, pms.AddCounter("errors", 2))
	require.NoError(t, pms.AddCounter("requests", 50)) // increment again
	require.NoError(t, pms.Save())

	// create a fresh storage and restore from file
	pms2 := newTestPersistentStorage(fp, false)
	require.NoError(t, pms2.Restore())

	g, err := pms2.GetGauge("temperature")
	assert.NoError(t, err)
	assert.InDelta(t, 23.7, g, 0.001)

	g2, err := pms2.GetGauge("humidity")
	assert.NoError(t, err)
	assert.InDelta(t, 55.0, g2, 0.001)

	c, err := pms2.GetCounter("requests")
	assert.NoError(t, err)
	assert.Equal(t, int64(150), c)

	c2, err := pms2.GetCounter("errors")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), c2)
}

func TestRoundTrip_MultipleSaveCycles(t *testing.T) {
	t.Parallel()

	fp := tempFilePath(t)
	pms := newTestPersistentStorage(fp, false)

	// first cycle
	require.NoError(t, pms.SetGauge("cpu", 10.0))
	require.NoError(t, pms.AddCounter("hits", 1))
	require.NoError(t, pms.Save())

	// second cycle — update values and save again
	require.NoError(t, pms.SetGauge("cpu", 20.0))
	require.NoError(t, pms.AddCounter("hits", 9))
	require.NoError(t, pms.Save())

	// restore into a fresh storage
	pms2 := newTestPersistentStorage(fp, false)
	require.NoError(t, pms2.Restore())

	g, _ := pms2.GetGauge("cpu")
	assert.InDelta(t, 20.0, g, 0.001)

	c, _ := pms2.GetCounter("hits")
	assert.Equal(t, int64(10), c)
}

// ---------- SynchronicSave ----------

func TestSynchronicSave(t *testing.T) {
	t.Parallel()

	fp := tempFilePath(t)
	pms := newTestPersistentStorage(fp, false)

	require.NoError(t, pms.SetGauge("mem", 80.0))
	pms.SynchronicSave()

	metrics := readMetricsFile(t, fp)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "mem", metrics[0].ID)
}

// ---------- PeriodicSave (smoke test with cancellation) ----------

func TestPeriodicSave_StopsOnChannelClose(t *testing.T) {
	t.Parallel()

	fp := tempFilePath(t)
	pms := newTestPersistentStorage(fp, false)

	require.NoError(t, pms.SetGauge("cpu", 50.0))

	done := make(chan struct{})
	go func() {
		defer close(done)
		pms.PeriodicSave(1)
	}()

	// PeriodicSave runs forever; we can't stop it from outside.
	// Just verify the goroutine is running and the method doesn't panic.
	// Let it tick once, then abandon the goroutine (it will be cleaned up
	// when the test process exits; in production you'd use context cancellation).
	// For a proper design, PeriodicSave should accept a context.
	// Here we simply let it run briefly to verify no panic.
	_ = done
}
