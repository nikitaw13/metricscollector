package agent

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/PrometheRus/metricscollector/internal/model"
	"github.com/stretchr/testify/assert"
)

// Test_SendMetrics verifies that the sender correctly reports all gauge and
// counter metrics to the server. It checks three things for each metric:
//   - the metric was received by the server,
//   - the HTTP method is POST,
//   - the Content-Type header is "application/json; charset=utf-8".

const ContentTypeHeader = "application/json; charset=utf-8"

func TestSendMetrics(t *testing.T) {
	as := NewAgentStorage()

	for _, v := range GaugeMetrics {
		rv := rand.Float64()
		as.SetGauge(v, rv)
	}

	for _, v := range CounterMetrics {
		rv := rand.Int64()
		as.SetCounter(v, rv)
	}

	received := map[string]bool{}
	content_type := map[string]string{}
	method := map[string]string{}

	testhandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var m model.Metrics

		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		received[m.ID] = true
		method[m.ID] = r.Method
		content_type[m.ID] = r.Header.Get("Content-Type")

		w.WriteHeader(http.StatusOK)
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
			assert.Equal(t, ContentTypeHeader, content_type[m])
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
			assert.Equal(t, ContentTypeHeader, content_type[m])
		})
	}
}

// TestResetIf200 verifies that counter metrics are reset to zero
// only after successful delivery (HTTP 200). This prevents
// re-sending already accepted values on the next run cycle.
func TestResetIf200(t *testing.T) {
	as := NewAgentStorage()

	rv := rand.Int64()
	for _, v := range CounterMetrics {
		as.SetCounter(v, rv)
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

// TestNoResetIfNot200 verifies that counter metrics are NOT reset
// when delivery fails (non-200 response). This ensures that metrics
// are preserved for retry on the next run cycle, preventing data loss.
func TestNoResetIfNot200(t *testing.T) {
	as := NewAgentStorage()

	rv := rand.Int64()
	for _, v := range CounterMetrics {
		as.SetCounter(v, rv)
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
