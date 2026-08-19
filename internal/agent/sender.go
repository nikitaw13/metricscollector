package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/PrometheRus/metricscollector/internal/model"
)

// Sender is responsible for sending collected metrics to the server
// via HTTP POST requests.
type Sender struct {
	URL     string
	Storage Storage
	Client  http.Client
}

// Run performs a one-shot send of all stored gauge and counter metrics to the server.
// Counter metrics are reset to zero only after a successful (HTTP 200) response.
func (s *Sender) Run() {
	updateURL := fmt.Sprintf("%s/update", s.URL)

	// Serialize and POST every gauge metric.
	for metricName, value := range s.Storage.GetAllGauges() {
		m := model.Metric{
			Type:  model.Gauge,
			ID:    metricName,
			Value: &value,
		}

		jsonBody, err := json.Marshal(&m)
		if err != nil {
			log.Println(err)
			continue
		}

		compressedBody, err := Compress(jsonBody)
		if err != nil {
			log.Println(err)
			continue
		}

		r, err := http.NewRequest(
			http.MethodPost,
			updateURL,
			bytes.NewReader(compressedBody),
		)

		if err != nil {
			log.Println(err)
			continue
		}

		r.Header.Set("Content-Type", "application/json; charset=utf-8")
		r.Header.Set("Content-Encoding", "gzip")
		resp, err := s.Client.Do(r)

		if err != nil {
			log.Println(err)
			continue
		}
		logRequest(resp, metricName, updateURL)
		drainAndCloseResponse(resp)
	}

	// Serialize and POST every counter metric; reset on 200.
	for metricName, value := range s.Storage.GetAllCounters() {
		m := model.Metric{
			Type:  model.Counter,
			ID:    metricName,
			Delta: &value,
		}

		jsonBody, err := json.Marshal(&m)
		if err != nil {
			log.Println(err)
			continue
		}

		compressedBody, err := Compress(jsonBody)
		if err != nil {
			log.Println(err)
			continue
		}

		r, err := http.NewRequest(
			http.MethodPost,
			updateURL,
			bytes.NewReader(compressedBody))

		if err != nil {
			log.Println(err)
			continue
		}

		r.Header.Set("Content-Type", "application/json; charset=utf-8")
		r.Header.Set("Content-Encoding", "gzip")
		resp, err := s.Client.Do(r)
		if err != nil {
			log.Println(err)
			continue
		}
		resetCounter(resp, metricName, s.Storage)
		logRequest(resp, metricName, updateURL)
		drainAndCloseResponse(resp)
	}
}

// resetCounter zeroes the named counter after a successful (HTTP 200) send.
func resetCounter(resp *http.Response, m string, s Storage) {
	// Reset the counter to zero if 200
	if resp.StatusCode == http.StatusOK {
		s.ResetCounter(m)
	}
}

// logRequest logs the outcome of sending a single metric.
func logRequest(resp *http.Response, m string, url string) {
	if resp.StatusCode == http.StatusOK {
		log.Printf("Metric '%s' sent to %s", m, url)
	} else {
		log.Printf("Metric '%s' failed to send to %s, status: %d", m, url, resp.StatusCode)
	}
}

// drainAndCloseResponse drains and closes the response body to allow TCP connection reuse.
func drainAndCloseResponse(resp *http.Response) {
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
}
