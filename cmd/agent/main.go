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
	// Отправлять метрики на сервер с указанным URL
	url = "http://localhost:8080"
)

func main() {
	storage := agent.NewAgentStorage()
	collector := &agent.Collector{Storage: storage}
	sender := &agent.Sender{
		URL:     url,
		Storage: storage,
		Client: http.Client{
			Timeout: 5 * time.Second,
		},
	}

	// Два интервала в одном потоке через Sleep не реализовать
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
