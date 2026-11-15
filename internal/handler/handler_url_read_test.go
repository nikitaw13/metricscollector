package handler

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMetricsHandler_handleURLRead verifies the metric list page and single metric value reads.
func TestMetricsHandler_handleURLRead(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		want   expectedResponse
	}{
		{
			"Read all metrics", http.MethodGet, "/", expectedResponse{
				http.StatusOK,
				expectedHTMLResponse,
				"text/html; charset=UTF-8",
			},
		},
		{
			"Read existing gauge", http.MethodGet, "/value/gauge/___test___", expectedResponse{
				statusCode:  http.StatusOK,
				body:        "123\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			"Read existing counter", http.MethodGet, "/value/counter/___test___", expectedResponse{
				statusCode:  http.StatusOK,
				body:        "234\n",
				contentType: "text/plain; charset=utf-8",
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
