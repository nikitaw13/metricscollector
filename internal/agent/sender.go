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
			Type: model.Gauge,
			ID:    metricName,
			Value: &value,
		}

		jsonBody, err := json.Marshal(&m)

		if err != nil {
			log.Println(err)
			continue
		}

		req, err := http.NewRequest(
			http.MethodPost,
			updateURL,
			bytes.NewReader(jsonBody),
		)

		if err != nil {
			log.Println(err)
			continue
		}

		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		resp, err := s.Client.Do(req)

		if err != nil {
			log.Println(err)
			continue
		}
		logRequest(resp, metricName, updateURL)
		finalizeSend(resp)
	}

	// Serialize and POST every counter metric; reset on 200.
	for metricName, value := range s.Storage.GetAllCounters() {
		m := model.Metric{
			Type: model.Counter,
			ID:    metricName,
			Delta: &value,
		}

		jsonBody, err := json.Marshal(&m)

		if err != nil {
			log.Println(err)
			continue
		}

		req, err := http.NewRequest(http.MethodPost, updateURL, bytes.NewReader(jsonBody))

		if err != nil {
			log.Println(err)
			continue
		}

		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		resp, err := s.Client.Do(req)
		if err != nil {
			log.Println(err)
			continue
		}
		resetCounter(resp, metricName, s.Storage)
		logRequest(resp, metricName, updateURL)
		finalizeSend(resp)
	}
}

// resetCounter zeroes the counter in local storage only after a 200 OK response.
func resetCounter(resp *http.Response, metricName string, storage Storage) {
	if resp.StatusCode == http.StatusOK {
		storage.ResetCounter(metricName)
	}
}

// logRequest prints whether a metric delivery succeeded or failed.
func logRequest(resp *http.Response, metricName string, url string) {
	if resp.StatusCode == http.StatusOK {
		log.Printf("Metric '%s' sent to %s", metricName, url)
	} else {
		log.Printf("Metric '%s' failed to send to %s, status: %d", metricName, url, resp.StatusCode)
	}
}

// finalizeSend drains the response body to allow TCP connection reuse (keep-alive).
func finalizeSend(resp *http.Response) {
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
}