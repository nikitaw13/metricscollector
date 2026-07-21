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

	collectorMonitor := agent.NewCollectorMonitor(time.Second * pollInterval)
	senderMonitor := agent.NewSenderMonitor(time.Second * reportInterval)

	collector := &agent.Collector{Storage: storage}
	sender := &agent.Sender{Storage: storage}

	collectorMonitor.Run(collector)
	senderMonitor.Run(sender)
}
