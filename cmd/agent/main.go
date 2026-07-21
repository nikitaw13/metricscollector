package main

import (
	"time"

	"github.com/PrometheRus/metricscollector/internal/agent"
)

const (
	// Обновлять метрики из пакета runtime с заданной частотой
	pollInterval = 2
	// Отправлять метрики на сервер с заданной частотой
	reportInterval = 10
)

func main() {
	storage := agent.NewAgentStorage()
	collector := &agent.Collector{Storage: storage}
	sender := &agent.Sender{Storage: storage}

	tick := 0
	for {
		time.Sleep(1 * time.Second)
		tick++
		if tick%pollInterval == 0 {
			collector.Run()
		}
		if tick%reportInterval == 0 {
			sender.Run()
		}
	}
}
