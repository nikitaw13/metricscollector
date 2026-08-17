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
	for metric, value := range s.Storage.GetAllGauges() {
		m := model.Metrics{
			MType: model.Gauge,
			ID:    metric,
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
		res, err := s.Client.Do(req)

		if err != nil {
			log.Println(err)
			continue
		}
		logRequest(res, metric, updateURL)
		finalizeSend(res)
	}

	// Serialize and POST every counter metric; reset on 200.
	for metric, value := range s.Storage.GetAllCounters() {
		m := model.Metrics{
			MType: model.Counter,
			ID:    metric,
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
		res, err := s.Client.Do(req)
		if err != nil {
			log.Println(err)
			continue
		}
		resetCounter(res, metric, s.Storage)
		logRequest(res, metric, updateURL)
		finalizeSend(res)
	}
}

// resetCounter zeroes the counter in local storage only after a 200 OK response.
func resetCounter(res *http.Response, m string, s Storage) {
	if res.StatusCode == http.StatusOK {
		s.ResetCounter(m)
	}
}

// logRequest prints whether a metric delivery succeeded or failed.
func logRequest(res *http.Response, m string, fp string) {
	if res.StatusCode == http.StatusOK {
		log.Printf("Metric '%s' sent to %s", m, fp)
	} else {
		log.Printf("Metric '%s' failed to send to %s, status: %d", m, fp, res.StatusCode)
	}
}

// Drain response body to allow TCP connection reuse (keep-alive).
func finalizeSend(res *http.Response) {
	io.Copy(io.Discard, res.Body)
	res.Body.Close()
}
