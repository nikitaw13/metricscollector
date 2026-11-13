package handler

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricsHandler_handleURLUpdate(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		want   expectedResponse
	}{{
		"Counter positive int", http.MethodPost, "/update/counter/test/1000", expectedResponse{
			http.StatusOK,
			"Metric 'test' updated\n",
			"text/plain; charset=utf-8",
		},
	},
		{
			"Counter negative int", http.MethodPost, "/update/counter/test/-1001", expectedResponse{
				http.StatusOK,
				"Metric 'test' updated\n",
				"text/plain; charset=utf-8",
			},
		},
		{
			"Gauge positive int", http.MethodPost, "/update/gauge/test/1000", expectedResponse{
				http.StatusOK,
				"Metric 'test' updated\n",
				"text/plain; charset=utf-8",
			},
		},
		{
			"Gauge negative int", http.MethodPost, "/update/gauge/test/-1000", expectedResponse{
				http.StatusOK,
				"Metric 'test' updated\n",
				"text/plain; charset=utf-8",
			},
		},
		{
			"Gauge positive float", http.MethodPost, "/update/gauge/test/1000.00", expectedResponse{
				http.StatusOK,
				"Metric 'test' updated\n",
				"text/plain; charset=utf-8",
			},
		},
		{
			"Gauge negative float", http.MethodPost, "/update/gauge/test/-1000.00", expectedResponse{
				http.StatusOK,
				"Metric 'test' updated\n",
				"text/plain; charset=utf-8",
			},
		}}

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
