package main

import (
	"net/http"
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
	sender := &agent.Sender{Storage: storage, Client: http.Client{Timeout: 5 * time.Second}}

	go func() {
		for {
			collector.Run()
			time.Sleep(pollInterval * time.Second)
		}
	}()

	for {
		sender.Run()
		time.Sleep(reportInterval * time.Second)
	}
}
