package handler

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/nikitaw13/metricscollector/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMetricsHandler_handleJSONRead verifies that /value returns the correct stored metric values in JSON.
func TestMetricsHandler_handleJSONRead(t *testing.T) {
	tests := []struct {
		name string
		body string
		want model.Metric
	}{
		{
			"Read existing gauge",
			"{\"type\":\"gauge\", \"id\":\"___test___\"}",
			model.Metric{ID: "___test___", Type: "gauge", Value: ptrFloat64(defaultGaugeValue)},
		},
		{
			"Read existing counter",
			"{\"type\":\"counter\", \"id\":\"___test___\"}",
			model.Metric{ID: "___test___", Type: "counter", Delta: ptrInt64(defaultCounterValue)},
		},
	}

	ts := newTestServer()
	t.Cleanup(ts.Close)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resp := doJSONRequest(t, ts, http.MethodPost, "/value", tc.body, "application/json")
			assert.Equal(t, http.StatusOK, resp.statusCode)
			assert.Equal(t, expectedJSONContentType, resp.header.Get("Content-Type"))

			var got model.Metric
			require.NoError(t, json.Unmarshal([]byte(resp.body), &got))
			assert.Equal(t, tc.want, got)
		})
	}
}
