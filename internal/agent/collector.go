package agent

import (
	"fmt"
	"log"
	"math/rand/v2"
	"runtime"

	"github.com/nikitaw13/metricscollector/internal/model"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// CollectRuntimeMetrics gathers one snapshot of runtime memory statistics and custom metrics and returns them as a batch.
func CollectRuntimeMetrics() []model.Metric {
	var storage = NewAgentStorage()
	var metrics []model.Metric
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	storage.SetGauge("Alloc", float64(memStats.Alloc))
	storage.SetGauge("BuckHashSys", float64(memStats.BuckHashSys))
	storage.SetGauge("Frees", float64(memStats.Frees))
	storage.SetGauge("GCCPUFraction", float64(memStats.GCCPUFraction))
	storage.SetGauge("GCSys", float64(memStats.GCSys))
	storage.SetGauge("HeapAlloc", float64(memStats.HeapAlloc))
	storage.SetGauge("HeapIdle", float64(memStats.HeapIdle))
	storage.SetGauge("HeapInuse", float64(memStats.HeapInuse))
	storage.SetGauge("HeapObjects", float64(memStats.HeapObjects))
	storage.SetGauge("HeapReleased", float64(memStats.HeapReleased))
	storage.SetGauge("HeapSys", float64(memStats.HeapSys))
	storage.SetGauge("LastGC", float64(memStats.LastGC))
	storage.SetGauge("Lookups", float64(memStats.Lookups))
	storage.SetGauge("MCacheInuse", float64(memStats.MCacheInuse))
	storage.SetGauge("MCacheSys", float64(memStats.MCacheSys))
	storage.SetGauge("MSpanInuse", float64(memStats.MSpanInuse))
	storage.SetGauge("MSpanSys", float64(memStats.MSpanSys))
	storage.SetGauge("Mallocs", float64(memStats.Mallocs))
	storage.SetGauge("NextGC", float64(memStats.NextGC))
	storage.SetGauge("NumForcedGC", float64(memStats.NumForcedGC))
	storage.SetGauge("NumGC", float64(memStats.NumGC))
	storage.SetGauge("OtherSys", float64(memStats.OtherSys))
	storage.SetGauge("PauseTotalNs", float64(memStats.PauseTotalNs))
	storage.SetGauge("StackInuse", float64(memStats.StackInuse))
	storage.SetGauge("StackSys", float64(memStats.StackSys))
	storage.SetGauge("Sys", float64(memStats.Sys))
	storage.SetGauge("TotalAlloc", float64(memStats.TotalAlloc))
	// Custom Metrics
	storage.AddCounter("PollCount", 1)                  // increments by 1 on each collection cycle
	storage.SetGauge("RandomValue", rand.NormFloat64()) // random normally-distributed value

	for key, delta := range storage.DrainCounters() {
		var metric model.Metric
		metric.ID = key
		metric.Delta = &delta
		metric.Type = model.Counter
		metrics = append(metrics, metric)
	}

	for key, value := range storage.GetAllGauges() {
		var metric model.Metric
		metric.ID = key
		metric.Value = &value
		metric.Type = model.Gauge
		metrics = append(metrics, metric)
	}

	return metrics
}

// CollectSystemMetrics gathers one snapshot of system metrics via gopsutil: total and free memory, plus per-core CPU utilization, and returns them as a batch.
func CollectSystemMetrics() []model.Metric {
	var storage = NewAgentStorage()
	var metrics []model.Metric

	memoryStat, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("error getting memory stats: %v", err)
		return nil
	}

	cpuUtilization, err := cpu.Percent(0, true)
	if err != nil {
		log.Printf("error getting CPU utilization: %v", err)
		return nil
	}

	storage.SetGauge("TotalMemory", float64(memoryStat.Total))
	storage.SetGauge("FreeMemory", float64(memoryStat.Free))

	for i, utilization := range cpuUtilization {
		storage.SetGauge(fmt.Sprintf("CPUutilization%d", i+1), float64(utilization))
	}

	for key, value := range storage.GetAllGauges() {
		var metric model.Metric
		metric.ID = key
		metric.Value = &value
		metric.Type = model.Gauge
		metrics = append(metrics, metric)
	}
	return metrics
}
