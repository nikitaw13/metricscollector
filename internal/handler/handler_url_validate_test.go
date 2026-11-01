package handler

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestURLValidate verifies that requests with an invalid type, name, or value are rejected with 4xx.
func TestURLValidate(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		want   expectedResponse
	}{
		// Missing or invalid metric type
		{
			"Missing Content-Type header", http.MethodPost, "/update", expectedResponse{
				http.StatusBadRequest,
				"Type is required\n",
				"text/plain; charset=utf-8",
			},
		},
		{
			"Invalid metric type", http.MethodPost, "/update/random/", expectedResponse{
				http.StatusBadRequest,
				"Invalid metric type\n",
				"text/plain; charset=utf-8",
			},
		},

		// Missing or invalid metric name
		{
			"Missing gauge metric name", http.MethodPost, "/update/gauge", expectedResponse{
				http.StatusNotFound,
				"Metric ID is required\n",
				"text/plain; charset=utf-8",
			},
		},
		{
			"Missing counter metric name", http.MethodPost, "/update/counter", expectedResponse{
				http.StatusNotFound,
				"Metric ID is required\n",
				"text/plain; charset=utf-8",
			},
		},
		{
			"Invalid gauge metric name", http.MethodGet, "/value/gauge/unknown", expectedResponse{
				http.StatusNotFound,
				"gauge unknown not found\n",
				"text/plain; charset=utf-8",
			},
		},
		{
			"Invalid counter metric name", http.MethodGet, "/value/counter/unknown", expectedResponse{
				http.StatusNotFound,
				"counter unknown not found\n",
				"text/plain; charset=utf-8",
			},
		},

		// Missing or invalid metric value
		{
			"Missing gauge metric value", http.MethodPost, "/update/gauge/name", expectedResponse{
				http.StatusBadRequest,
				"Metric value is required\n",
				"text/plain; charset=utf-8",
			},
		},
		{
			"Missing counter metric value", http.MethodPost, "/update/counter/name", expectedResponse{
				http.StatusBadRequest,
				"Metric value is required\n",
				"text/plain; charset=utf-8",
			},
		},
		{
			"Gauge metric value is non-numeric", http.MethodPost, "/update/gauge/name/value", expectedResponse{
				http.StatusBadRequest,
				"Invalid metric value\n",
				"text/plain; charset=utf-8",
			},
		},
		{
			"Counter metric value is non-numeric", http.MethodPost, "/update/counter/name/value", expectedResponse{
				http.StatusBadRequest,
				"Invalid metric value\n",
				"text/plain; charset=utf-8",
			},
		},
		{
			"Counter metric value is positive float", http.MethodPost, "/update/counter/name/1000.00", expectedResponse{
				http.StatusBadRequest,
				"Invalid metric value\n",
				"text/plain; charset=utf-8",
			},
		},
		{
			"Counter metric value is negative float", http.MethodPost, "/update/counter/name/-1001,00", expectedResponse{
				http.StatusBadRequest,
				"Invalid metric value\n",
				"text/plain; charset=utf-8",
			},
		},
	}

	ts := newTestServer()
	t.Cleanup(ts.Close)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resp := doRequest(t, ts, tc.method, tc.path)
			assert.Equal(t, tc.want.statusCode, resp.statusCode)
			assert.Equal(t, tc.want.body, resp.body)
			assert.Equal(t, tc.want.contentType, resp.header.Get("Content-Type"))
		})
	}
}
