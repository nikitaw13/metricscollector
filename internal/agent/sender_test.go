package agent

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestSendMetrics verifies that the sender correctly reports all gauge and
// counter metrics to the server. It checks three things for each metric:
//   - the metric was received by the server,
//   - the HTTP method is POST,
//   - the Content-Type header is "text/plain; charset=utf-8".
func TestSendMetrics(t *testing.T) {
	as := NewAgentStorage()

	for _, v := range GaugeMetrics {
		rv := rand.Float64()
		as.SetGauge(v, rv)
	}

	for _, v := range CounterMetrics {
		rv := rand.Int64()
		as.AddCounter(v, rv)
	}

	received := map[string]bool{}
	contentType := map[string]string{}
	method := map[string]string{}

	testhandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metric := strings.Split(r.URL.Path, "/")[3]
		received[metric] = true
		method[metric] = r.Method
		contentType[metric] = r.Header.Get("Content-Type")
	})

	ts := httptest.NewServer(testhandler)
	defer ts.Close()

	sender := &Sender{
		URL:     ts.URL,
		Storage: as,
		Client: http.Client{
			Timeout: 5 * time.Second,
		},
	}

	sender.Run()

	for _, m := range GaugeMetrics {
		t.Run(fmt.Sprintf("Received %v", m), func(t *testing.T) {
			assert.Equal(t, true, received[m])
		})
		t.Run(fmt.Sprintf("Method %v", m), func(t *testing.T) {
			assert.Equal(t, http.MethodPost, method[m])
		})
		t.Run(fmt.Sprintf("Content-Type %v", m), func(t *testing.T) {
			assert.Equal(t, "text/plain; charset=utf-8", contentType[m])
		})
	}
	for _, m := range CounterMetrics {
		t.Run(fmt.Sprintf("Received %v", m), func(t *testing.T) {
			assert.Equal(t, true, received[m])
		})
		t.Run(fmt.Sprintf("Method %v", m), func(t *testing.T) {
			assert.Equal(t, http.MethodPost, method[m])
		})
		t.Run(fmt.Sprintf("Content-Type %v", m), func(t *testing.T) {
			assert.Equal(t, "text/plain; charset=utf-8", contentType[m])
		})
	}
}

// Test_ResetIf200 verifies that all counter metrics are reset to zero
// after the server responds with HTTP 200 OK.
func TestResetIf200(t *testing.T) {
	as := NewAgentStorage()

	rv := rand.Int64()
	for _, v := range CounterMetrics {
		as.AddCounter(v, rv)
	}

	testhandler := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
	})

	ts := httptest.NewServer(testhandler)
	defer ts.Close()

	sender := &Sender{
		URL:     ts.URL,
		Storage: as,
		Client: http.Client{
			Timeout: 5 * time.Second,
		},
	}

	sender.Run()

	for _, m := range CounterMetrics {
		t.Run(fmt.Sprintf("Received %v", m), func(t *testing.T) {
			val, _ := as.GetCounter(m)
			assert.Equal(t, int64(0), val)
		})
	}
}

// Test_NoResetIfNot200 verifies that counter metrics are NOT reset
// when the server responds with a non-200 status code (e.g. HTTP 500).
func TestNoResetIfNot200(t *testing.T) {
	as := NewAgentStorage()

	rv := rand.Int64()
	for _, v := range CounterMetrics {
		as.AddCounter(v, rv)
	}

	testhandler := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusInternalServerError)
	})

	ts := httptest.NewServer(testhandler)
	defer ts.Close()

	sender := &Sender{
		URL:     ts.URL,
		Storage: as,
		Client: http.Client{
			Timeout: 5 * time.Second,
		},
	}

	sender.Run()

	for _, m := range CounterMetrics {
		t.Run(fmt.Sprintf("Received %v", m), func(t *testing.T) {
			val, _ := as.GetCounter(m)
			assert.Equal(t, rv, val)
		})
	}
}
