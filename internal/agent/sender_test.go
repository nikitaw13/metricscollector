package agent

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nikitaw13/metricscollector/internal/model"
	"github.com/stretchr/testify/assert"
)

// expectedContentType is the expected Content-Type header for all JSON requests from the agent.
const expectedContentType = "application/json; charset=utf-8"

// TestSendMetrics verifies that the sender correctly reports all gauge and
// counter metrics to the server. It checks three things for each metric:
//   - the metric was received by the server,
//   - the HTTP method is POST,
//   - the Content-Type header is "application/json; charset=utf-8".
func TestSendMetrics(t *testing.T) {
	storage := NewAgentStorage()

	for _, metricName := range GaugeMetrics {
		initialValue := rand.Float64()
		storage.SetGauge(metricName, initialValue)
	}

	for _, metricName := range CounterMetrics {
		initialValue := rand.Int64()
		storage.AddCounter(metricName, initialValue)
	}

	received := map[string]bool{}
	contentType := map[string]string{}
	method := map[string]string{}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var metric model.Metric

		if r.Header.Get("Content-Encoding") == "gzip" {

			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			defer gz.Close()
			r.Header.Del("Content-Encoding")
			r.Header.Del("Content-Length")
			r.ContentLength = -1
			r.Body = io.NopCloser(gz)

		}

		if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		received[metric.ID] = true
		method[metric.ID] = r.Method
		contentType[metric.ID] = r.Header.Get("Content-Type")

		w.WriteHeader(http.StatusOK)
	})

	ts := httptest.NewServer(testHandler)
	defer ts.Close()

	sender := &Sender{
		BaseURL: ts.URL,
		Storage: storage,
		Client: http.Client{
			Timeout: 5 * time.Second,
		},
	}

	sender.Run()

	for _, metricName := range GaugeMetrics {
		t.Run(fmt.Sprintf("Received %v", metricName), func(t *testing.T) {
			assert.Equal(t, true, received[metricName])
		})
		t.Run(fmt.Sprintf("Method %v", metricName), func(t *testing.T) {
			assert.Equal(t, http.MethodPost, method[metricName])
		})
		t.Run(fmt.Sprintf("Content-Type %v", metricName), func(t *testing.T) {
			assert.Equal(t, expectedContentType, contentType[metricName])
		})
	}
	for _, metricName := range CounterMetrics {
		t.Run(fmt.Sprintf("Received %v", metricName), func(t *testing.T) {
			assert.Equal(t, true, received[metricName])
		})
		t.Run(fmt.Sprintf("Method %v", metricName), func(t *testing.T) {
			assert.Equal(t, http.MethodPost, method[metricName])
		})
		t.Run(fmt.Sprintf("Content-Type %v", metricName), func(t *testing.T) {
			assert.Equal(t, expectedContentType, contentType[metricName])
		})
	}
}

// TestResetCounterOnSuccess verifies that counter metrics are reset to zero
// only after successful delivery (HTTP 200).
func TestResetCounterOnSuccess(t *testing.T) {
	storage := NewAgentStorage()

	initialValue := rand.Int64()
	for _, metricName := range CounterMetrics {
		storage.AddCounter(metricName, initialValue)
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ts := httptest.NewServer(testHandler)
	defer ts.Close()

	sender := &Sender{
		BaseURL: ts.URL,
		Storage: storage,
		Client: http.Client{
			Timeout: 5 * time.Second,
		},
	}

	sender.Run()

	for _, metricName := range CounterMetrics {
		t.Run(fmt.Sprintf("counter %v reset to zero", metricName), func(t *testing.T) {
			value, _ := storage.GetCounter(metricName)
			assert.Equal(t, int64(0), value)
		})
	}
}

// TestKeepCounterOnError verifies that counter metrics are NOT reset
// when delivery fails (non-200 response), preserving them for retry.
func TestKeepCounterOnError(t *testing.T) {
	storage := NewAgentStorage()

	initialValue := rand.Int64()
	for _, metricName := range CounterMetrics {
		storage.AddCounter(metricName, initialValue)
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	ts := httptest.NewServer(testHandler)
	defer ts.Close()

	sender := &Sender{
		BaseURL: ts.URL,
		Storage: storage,
		Client: http.Client{
			Timeout: 5 * time.Second,
		},
	}

	sender.Run()

	for _, metricName := range CounterMetrics {
		t.Run(fmt.Sprintf("counter %v preserved on error", metricName), func(t *testing.T) {
			value, _ := storage.GetCounter(metricName)
			assert.Equal(t, initialValue, value)
		})
	}
}
