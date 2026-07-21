package agent

import (
	"log"
	"math/rand/v2"
	"runtime"
)

type Collector struct {
	Storage Storage
}

// One-shot run
func (c *Collector) Run() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	c.Storage.UpdateGauge("Alloc", float64(memStats.Alloc))
	c.Storage.UpdateGauge("BuckHashSys", float64(memStats.BuckHashSys))
	c.Storage.UpdateGauge("Frees", float64(memStats.Frees))
	c.Storage.UpdateGauge("GCCPUFraction", float64(memStats.GCCPUFraction))
	c.Storage.UpdateGauge("GCSys", float64(memStats.GCSys))
	c.Storage.UpdateGauge("HeapAlloc", float64(memStats.HeapAlloc))
	c.Storage.UpdateGauge("HeapIdle", float64(memStats.HeapIdle))
	c.Storage.UpdateGauge("HeapInuse", float64(memStats.HeapInuse))
	c.Storage.UpdateGauge("HeapObjects", float64(memStats.HeapObjects))
	c.Storage.UpdateGauge("HeapReleased", float64(memStats.HeapReleased))
	c.Storage.UpdateGauge("HeapSys", float64(memStats.HeapSys))
	c.Storage.UpdateGauge("LastGC", float64(memStats.LastGC))
	c.Storage.UpdateGauge("Lookups", float64(memStats.Lookups))
	c.Storage.UpdateGauge("MCacheInuse", float64(memStats.MCacheInuse))
	c.Storage.UpdateGauge("MCacheSys", float64(memStats.MCacheSys))
	c.Storage.UpdateGauge("MSpanInuse", float64(memStats.MSpanInuse))
	c.Storage.UpdateGauge("MSpanSys", float64(memStats.MSpanSys))
	c.Storage.UpdateGauge("Mallocs", float64(memStats.Mallocs))
	c.Storage.UpdateGauge("NextGC", float64(memStats.NextGC))
	c.Storage.UpdateGauge("NumForcedGC", float64(memStats.NumForcedGC))
	c.Storage.UpdateGauge("NumGC", float64(memStats.NumGC))
	c.Storage.UpdateGauge("OtherSys", float64(memStats.OtherSys))
	c.Storage.UpdateGauge("PauseTotalNs", float64(memStats.PauseTotalNs))
	c.Storage.UpdateGauge("StackInuse", float64(memStats.StackInuse))
	c.Storage.UpdateGauge("StackSys", float64(memStats.StackSys))
	c.Storage.UpdateGauge("Sys", float64(memStats.Sys))
	c.Storage.UpdateGauge("TotalAlloc", float64(memStats.TotalAlloc))
	// Custom Metrics
	c.Storage.UpdateCounter("PollCount", 1)                  //счётчик, увеличивающийся на 1 при каждом обновлении метрики из пакета runtime
	c.Storage.UpdateGauge("RandomValue", rand.NormFloat64()) // обновляемое произвольное значение
	log.Print("The metrics have been harvested by collector")
}
