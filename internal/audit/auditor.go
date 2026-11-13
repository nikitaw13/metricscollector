// Package audit publishes metric-update events to the configured receivers.
package audit

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/nikitaw13/metricscollector/internal/model"
)

// Auditor fans out metric-update events to the configured receivers.
type Auditor struct {
	receivers []subscriber
	mu        sync.Mutex
}

// event is the JSON payload that Notify publishes to all receivers.
type event struct {
	Timestamp   int64    `json:"ts"`
	MetricNames []string `json:"metrics"`
	IPAddress   string   `json:"ip_address"`
}

// Close closes all registered receivers, returning their errors joined.
func (a *Auditor) Close() error {
	var snapshot []subscriber

	a.mu.Lock()
	snapshot = slices.Clone(a.receivers)
	a.mu.Unlock()

	var merged error
	for _, receiver := range snapshot {
		err := receiver.close()
		if err != nil {
			merged = errors.Join(merged, err)
		}
	}
	return merged
}

// New creates an Auditor that writes events to file and/or url when set, using client for URL delivery. It reports an error if the audit file cannot be opened.
func New(fileName, url string, client httpClient) (*Auditor, error) {
	a := &Auditor{}
	if fileName != "" {
		newFile, err := openAuditFile(fileName)
		if err != nil {
			return a, err
		}
		a.subscribe(newFileWriter(newFile))
	}

	if url != "" && client != nil {
		a.subscribe(newURLSender(url, client))
	}
	return a, nil
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
	for _, err := range sendErrs {
		if err != nil {
			return fmt.Errorf("failed to send an event: %w", errors.Join(sendErrs...))
		}
	}
	return nil
}

// notify sends event to all receivers in parallel, returning one result per receiver.
func (a *Auditor) notify(event []byte) []error {
	var wg sync.WaitGroup
	var snapshot []subscriber

	a.mu.Lock()
	snapshot = slices.Clone(a.receivers)
	a.mu.Unlock()

	var sendErrs = make([]error, len(snapshot))

	for i, receiver := range snapshot {
		wg.Go(func() {
			sendErrs[i] = receiver.send(event)
		})

	}
	wg.Wait()
	return sendErrs
}

// subscribe registers s as a receiver of future events.
func (a *Auditor) subscribe(s subscriber) {
	a.mu.Lock()
	a.receivers = append(a.receivers, s)
	a.mu.Unlock()
}

// unsubscribe removes s from the receivers; it currently has no callers and
// is kept to complete the observer pattern.
func (a *Auditor) unsubscribe(s subscriber) {
	a.mu.Lock()
	for i, receiver := range a.receivers {
		if receiver == s {
			a.receivers = append(a.receivers[:i], a.receivers[i+1:]...)
			break
		}

	}
	a.mu.Unlock()
}
