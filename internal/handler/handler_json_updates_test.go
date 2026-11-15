package handler

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/nikitaw13/metricscollector/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMetricsHandler_handleJSONUpdates verifies that POST /updates accepts valid metric batches
// and rejects invalid payloads with 400.
func TestMetricsHandler_handleJSONUpdates(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		path        string
		body        string
		contentType string
		want        expectedResponse
	}{
		{
			"Valid batch with gauge and counter",
			http.MethodPost,
			"/updates/",
			`[{"type":"gauge","id":"batch_gauge","value":42.5},{"type":"counter","id":"batch_counter","delta":7}]`,
			"application/json",
			expectedResponse{http.StatusOK, "", ""},
		},
		{
			"Empty batch is accepted",
			http.MethodPost,
			"/updates/",
			`[]`,
			"application/json",
			expectedResponse{http.StatusOK, "", ""},
		},
		{
			"Batch with invalid metric type",
			http.MethodPost,
			"/updates/",
			`[{"type":"gauge","id":"batch_gauge","value":1.0},{"type":"random","id":"batch_unknown"}]`,
			"application/json",
			expectedResponse{http.StatusBadRequest, "", expectedJSONContentType},
		},
		{
			"Batch with missing gauge value",
			http.MethodPost,
			"/updates/",
			`[{"type":"gauge","id":"batch_gauge"}]`,
			"application/json",
			expectedResponse{http.StatusBadRequest, "", expectedJSONContentType},
		},
		{
			"Batch with missing metric name",
			http.MethodPost,
			"/updates/",
			`[{"type":"counter","delta":5}]`,
			"application/json",
			expectedResponse{http.StatusBadRequest, "", expectedJSONContentType},
		},
		{
			"Non-array JSON body",
			http.MethodPost,
			"/updates/",
			`{"type":"gauge","id":"batch_gauge","value":1.0}`,
			"application/json",
			expectedResponse{http.StatusBadRequest, "", expectedJSONContentType},
		},
		{
			"Invalid JSON syntax",
			http.MethodPost,
			"/updates/",
			`[{invalid json`,
			"application/json",
			expectedResponse{http.StatusBadRequest, "", expectedJSONContentType},
		},
		{
			"Empty request body",
			http.MethodPost,
			"/updates/",
			"",
			"application/json",
			expectedResponse{http.StatusBadRequest, "", expectedJSONContentType},
		},
	}

	ts := newTestServer()
	t.Cleanup(ts.Close)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resp := doJSONRequest(t, ts, tc.method, tc.path, tc.body, tc.contentType)
			assert.Equal(t, tc.want.statusCode, resp.statusCode)
			// Success responses carry no body, so no Content-Type is set.
			if tc.want.contentType != "" {
				assert.Equal(t, tc.want.contentType, resp.header.Get("Content-Type"))
			}
		})
	}
}

// TestMetricsHandler_handleJSONUpdatesPersistsBatch verifies that a valid batch is fully stored
// and readable via POST /value.
func TestMetricsHandler_handleJSONUpdatesPersistsBatch(t *testing.T) {
	ts := newTestServer()
	t.Cleanup(ts.Close)

	batch := `[{"type":"gauge","id":"batch_gauge","value":42.5},{"type":"counter","id":"batch_counter","delta":7}]`
	resp := doJSONRequest(t, ts, http.MethodPost, "/updates/", batch, "application/json")
	require.Equal(t, http.StatusOK, resp.statusCode)

	tests := []struct {
		name string
		body string
		want model.Metric
	}{
		{
			"Read batch gauge",
			`{"type":"gauge","id":"batch_gauge"}`,
			model.Metric{ID: "batch_gauge", Type: "gauge", Value: ptrFloat64(42.5)},
		},
		{
			"Read batch counter",
			`{"type":"counter","id":"batch_counter"}`,
			model.Metric{ID: "batch_counter", Type: "counter", Delta: ptrInt64(7)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resp := doJSONRequest(t, ts, http.MethodPost, "/value", tc.body, "application/json")
			require.Equal(t, http.StatusOK, resp.statusCode)

			var got model.Metric
			require.NoError(t, json.Unmarshal([]byte(resp.body), &got))
			assert.Equal(t, tc.want, got)
		})
	}
}

// TestMetricsHandler_handleJSONUpdatesRejectsBatchAtomically verifies that a batch containing an
// invalid metric is rejected entirely: no metric from the batch is stored.
func TestMetricsHandler_handleJSONUpdatesRejectsBatchAtomically(t *testing.T) {
	ts := newTestServer()
	t.Cleanup(ts.Close)

	batch := `[{"type":"gauge","id":"atomic_gauge","value":1.5},{"type":"random","id":"atomic_unknown"}]`
	resp := doJSONRequest(t, ts, http.MethodPost, "/updates/", batch, "application/json")
	require.Equal(t, http.StatusBadRequest, resp.statusCode)

	// The valid metric from the rejected batch must not be stored.
	resp = doJSONRequest(t, ts, http.MethodPost, "/value", `{"type":"gauge","id":"atomic_gauge"}`, "application/json")
	require.Equal(t, http.StatusNotFound, resp.statusCode)
}
