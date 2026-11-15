package handler

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMetricsHandler_handleJSONUpdate verifies that /update accepts valid metric payloads and returns 200.
func TestMetricsHandler_handleJSONUpdate(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		path        string
		body        string
		contentType string
		want        expectedResponse
	}{
		{
			"Counter positive int",
			http.MethodPost,
			"/update",
			"{\"type\":\"counter\", \"id\":\"test\", \"delta\":1000}",
			"application/json",
			expectedResponse{
				http.StatusOK,
				"",
				expectedJSONContentType,
			},
		},
		{
			"Gauge positive int",
			http.MethodPost,
			"/update",
			`{"type":"gauge", "id":"test", "value":1000}`,
			"application/json",
			expectedResponse{
				http.StatusOK,
				"",
				expectedJSONContentType,
			},
		},
		{
			"Gauge negative int",
			http.MethodPost,
			"/update",
			`{"type":"gauge", "id":"test", "value":-1000}`,
			"application/json",
			expectedResponse{
				http.StatusOK,
				"",
				expectedJSONContentType,
			},
		},
		{
			"Gauge positive float",
			http.MethodPost,
			"/update",
			`{"type":"gauge", "id":"test", "value":1000.00}`,
			"application/json",
			expectedResponse{
				http.StatusOK,
				"",
				expectedJSONContentType,
			},
		},
		{
			"Gauge negative float",
			http.MethodPost,
			"/update",
			`{"type":"gauge", "id":"test", "value":-1000.00}`,
			"application/json",
			expectedResponse{
				http.StatusOK,
				"",
				expectedJSONContentType,
			},
		},
	}

	ts := newTestServer()
	t.Cleanup(ts.Close)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resp := doJSONRequest(t, ts, tc.method, tc.path, tc.body, tc.contentType)
			assert.Equal(t, tc.want.statusCode, resp.statusCode)
			assert.Equal(t, tc.want.contentType, resp.header.Get("Content-Type"))
		})
	}
}
