package agent

import (
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/nikitaw13/metricscollector/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// expectedContentType is the expected Content-Type header for all JSON requests from the agent.
const expectedContentType = "application/json; charset=utf-8"

// testHashKey is the HMAC key used by sender signing tests.
const testHashKey = "TestKey"

// TestSendMetrics verifies that the sender reports all gauge and counter
// metrics to the server in a single batched request. It checks:
//   - exactly one request is made per Run(),
//   - each metric was received by the server,
//   - the HTTP method is POST,
//   - the Content-Type header is "application/json; charset=utf-8".
func TestSendMetrics(t *testing.T) {
	t.Parallel()
	batch := Collect()

	var mu sync.Mutex
	received := map[string]bool{}
	contentType := map[string]string{}
	method := map[string]string{}
	requestCount := 0

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") != "gzip" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer gz.Close()

		var metrics []model.Metric
		if err := json.NewDecoder(gz).Decode(&metrics); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		mu.Lock()
		requestCount++
		for _, metric := range metrics {
			received[metric.ID] = true
			method[metric.ID] = r.Method
			contentType[metric.ID] = r.Header.Get("Content-Type")
		}
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
	})

	ts := httptest.NewServer(testHandler)
	defer ts.Close()

	retries := []time.Duration{1 * time.Millisecond, 3 * time.Millisecond, 5 * time.Millisecond}
	client := NewClientWithRetries(
		retries,
		&http.Client{Timeout: 5 * time.Second},
	)

	sender := NewSender(
		ts.URL,
		client,
		"",
	)

	sender.Run(batch)

	assert.Equal(t, 1, requestCount, "Run() must send exactly one batch request")

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

// retryTransport simulates transient network failures: the first `failures`
// attempts return a transport error, later attempts are delegated to the
// underlying transport. It records a timestamp for every attempt.
type retryTransport struct {
	failures int
	base     http.RoundTripper

	mu       sync.Mutex
	attempts []time.Time
}

// RoundTrip records the attempt and delegates it to the underlying transport.
func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.mu.Lock()
	t.attempts = append(t.attempts, time.Now())
	n := len(t.attempts)
	t.mu.Unlock()

	if n <= t.failures {
		return nil, fmt.Errorf("attempt %d: simulated connection reset", n)
	}
	return t.base.RoundTrip(req)
}

// TestRetrySucceedsAfterTransientFailures verifies that the sender retries
// transport errors and delivers the batch on a later attempt; the retried
// request must carry an intact gzip-compressed JSON body.
func TestRetrySucceedsAfterTransientFailures(t *testing.T) {
	t.Parallel()
	gaugeValue := 42.5
	counterDelta := int64(7)
	batch := []model.Metric{
		{ID: "retry_gauge", Type: model.Gauge, Value: &gaugeValue},
		{ID: "retry_counter", Type: model.Counter, Delta: &counterDelta},
	}

	var mu sync.Mutex
	received := map[string]bool{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer gz.Close()

		var metrics []model.Metric
		if err := json.NewDecoder(gz).Decode(&metrics); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		mu.Lock()
		for _, metric := range metrics {
			received[metric.ID] = true
		}
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	retries := []time.Duration{1 * time.Millisecond, 3 * time.Millisecond, 5 * time.Millisecond}
	rt := &retryTransport{failures: 2, base: http.DefaultTransport}
	client := NewClientWithRetries(
		retries,
		&http.Client{Timeout: 5 * time.Second, Transport: rt},
	)
	sender := NewSender(
		ts.URL,
		client,
		"",
	)

	sender.Run(batch)

	assert.Equal(t, 3, len(rt.attempts), "two failed attempts plus one successful retry expected")
	assert.True(t, received["retry_gauge"], "gauge must be delivered after retries")
	assert.True(t, received["retry_counter"], "counter must be delivered after retries")
}

// TestSendWithHash verifies that the sender sends the HMAC of the uncompressed JSON batch in the HashSHA256 header.
func TestSendWithHash(t *testing.T) {
	t.Parallel()
	gaugeValue := 123.45
	batch := []model.Metric{{ID: "testGauge", Type: model.Gauge, Value: &gaugeValue}}
	var gotHash string
	var plainBody []byte

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") != "gzip" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer gz.Close()

		plainBody, err = io.ReadAll(gz)
		if err != nil {
			t.Error("error while reading gz")
			return
		}

		gotHash = r.Header.Get("HashSHA256")
		w.WriteHeader(http.StatusOK)
	})

	ts := httptest.NewServer(testHandler)
	defer ts.Close()

	retries := []time.Duration{1 * time.Millisecond, 3 * time.Millisecond, 5 * time.Millisecond}
	client := NewClientWithRetries(
		retries,
		&http.Client{Timeout: 5 * time.Second},
	)

	sender := NewSender(
		ts.URL,
		client,
		testHashKey,
	)

	sender.Run(batch)

	hasher := hmac.New(sha256.New, []byte(testHashKey))
	hasher.Write(plainBody)
	assert.Equal(t, hex.EncodeToString(hasher.Sum(nil)), gotHash)

	expected, err := json.Marshal(batch)
	require.NoError(t, err)
	assert.JSONEq(t, string(expected), string(plainBody), "request body must be the exact JSON encoding of the batch")
}
