package agent

import (
	"log"
	"math/rand/v2"
	"runtime"
)

// Collector gathers runtime metrics and custom metrics into agent storage.
type Collector struct {
	Storage Storage
}

// Run collects one snapshot of runtime and custom metrics into storage.
func (c *Collector) Run() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	c.Storage.SetGauge("Alloc", float64(memStats.Alloc))
	c.Storage.SetGauge("BuckHashSys", float64(memStats.BuckHashSys))
	c.Storage.SetGauge("Frees", float64(memStats.Frees))
	c.Storage.SetGauge("GCCPUFraction", float64(memStats.GCCPUFraction))
	c.Storage.SetGauge("GCSys", float64(memStats.GCSys))
	c.Storage.SetGauge("HeapAlloc", float64(memStats.HeapAlloc))
	c.Storage.SetGauge("HeapIdle", float64(memStats.HeapIdle))
	c.Storage.SetGauge("HeapInuse", float64(memStats.HeapInuse))
	c.Storage.SetGauge("HeapObjects", float64(memStats.HeapObjects))
	c.Storage.SetGauge("HeapReleased", float64(memStats.HeapReleased))
	c.Storage.SetGauge("HeapSys", float64(memStats.HeapSys))
	c.Storage.SetGauge("LastGC", float64(memStats.LastGC))
	c.Storage.SetGauge("Lookups", float64(memStats.Lookups))
	c.Storage.SetGauge("MCacheInuse", float64(memStats.MCacheInuse))
	c.Storage.SetGauge("MCacheSys", float64(memStats.MCacheSys))
	c.Storage.SetGauge("MSpanInuse", float64(memStats.MSpanInuse))
	c.Storage.SetGauge("MSpanSys", float64(memStats.MSpanSys))
	c.Storage.SetGauge("Mallocs", float64(memStats.Mallocs))
	c.Storage.SetGauge("NextGC", float64(memStats.NextGC))
	c.Storage.SetGauge("NumForcedGC", float64(memStats.NumForcedGC))
	c.Storage.SetGauge("NumGC", float64(memStats.NumGC))
	c.Storage.SetGauge("OtherSys", float64(memStats.OtherSys))
	c.Storage.SetGauge("PauseTotalNs", float64(memStats.PauseTotalNs))
	c.Storage.SetGauge("StackInuse", float64(memStats.StackInuse))
	c.Storage.SetGauge("StackSys", float64(memStats.StackSys))
	c.Storage.SetGauge("Sys", float64(memStats.Sys))
	c.Storage.SetGauge("TotalAlloc", float64(memStats.TotalAlloc))
	// Custom metrics
	c.Storage.AddCounter("PollCount", 1)           // increments by 1 on each collector run
	c.Storage.SetGauge("RandomValue", rand.NormFloat64()) // arbitrary updated value
	log.Print("The metrics have been harvested by collector")
}
