package main

import (
	"time"

	"github.com/PrometheRus/metricscollector/internal/agent"
)

func main() {
	storage := agent.NewAgentStorage()

	collectorMonitor := agent.CollectorMonitor{Interval: time.Second * 2}
	senderMonitor := agent.SenderMonitor{Interval: time.Second * 10}

	collector := &agent.Collector{Storage: storage}
	sender := &agent.Sender{Storage: storage}

	collectorMonitor.Run(collector)
	senderMonitor.Run(sender)
}
