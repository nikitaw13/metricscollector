// Package audit publishes metric-update events to the configured receivers.
package audit

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/nikitaw13/metricscollector/internal/model"
)

// Auditor fans out metric-update events to the configured receivers.
type Auditor struct {
	receivers []subscriber
}

type event struct {
	Timestamp int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

// New creates an Auditor that writes events to file and/or url when set, using client for URL delivery.
func New(file, url string, client httpClient) *Auditor {
	a := &Auditor{}
	if file != "" {
		a.receivers = append(a.receivers, fileWriter{file: file})
	}
	if url != "" {
		a.receivers = append(a.receivers, urlSender{url: url, client: client})
	}
	return a
}

// Notify serializes the update into a JSON event and sends it to all receivers, returning any send errors joined.
func (a *Auditor) Notify(timestamp int64, metrics []model.Metric, ipAddress string) error {
	var metricsNames []string

	for i := range metrics {
		metricsNames = append(metricsNames, metrics[i].ID)
	}

	evt := event{timestamp, metricsNames, ipAddress}
	jsonEvent, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	sendErrs := a.notify(jsonEvent)
	if len(sendErrs) > 0 {
		return fmt.Errorf("failed to send an event: %w", errors.Join(sendErrs...))
	}

	return nil
}

func (a *Auditor) notify(event []byte) []error {
	var sendErrs []error
	for _, receiver := range a.receivers {
		err := receiver.send(event)
		if err != nil {
			sendErrs = append(sendErrs, err)
		}
	}
	return sendErrs
}
