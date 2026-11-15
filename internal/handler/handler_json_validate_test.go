package handler

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestJSONValidate verifies that /update and /value reject invalid JSON payloads.
func TestJSONValidate(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		path        string
		body        string
		contentType string
		want        expectedResponse
	}{
		// Missing or invalid metric TYPE
		{
			"Missing metric type",
			http.MethodPost,
			"/update",
			"{}",
			"application/json",
			expectedResponse{
				http.StatusBadRequest,
				"",
				expectedJSONContentType,
			},
		},
		{
			"Invalid metric type",
			http.MethodPost,
			"/update",
			"{\"type\":\"random\"}",
			"application/json",
			expectedResponse{
				http.StatusBadRequest,
				"",
				expectedJSONContentType,
			},
		},

		// Missing or invalid metric NAME
		{
			"Missing gauge metric name",
			http.MethodPost,
			"/update",
			"{\"type\":\"gauge\"}",
			"application/json",
			expectedResponse{
				http.StatusBadRequest,
				"",
				expectedJSONContentType,
			},
		},
		{
			"Missing counter metric name",
			http.MethodPost,
			"/update",
			"{\"type\":\"counter\"}",
			"application/json",
			expectedResponse{
				http.StatusBadRequest,
				"",
				expectedJSONContentType,
			},
		},
		{
			"Invalid gauge metric name",
			http.MethodPost,
			"/value",
			"{\"type\":\"gauge\",\"id\":\"unknown\"}",
			"application/json",
			expectedResponse{
				http.StatusNotFound,
				"",
				expectedJSONContentType,
			},
		},
		{
			"Invalid counter metric name",
			http.MethodPost,
			"/value",
			"{\"type\":\"counter\",\"id\":\"unknown\"}",
			"application/json",
			expectedResponse{
				http.StatusNotFound,
				"",
				expectedJSONContentType,
			},
		},
		// POST /value: invalid payloads (syntax, type, name)
		{
			"Value invalid JSON syntax",
			http.MethodPost,
			"/value",
			"{\"type\":\"gauge\", \"id\":",
			"application/json",
			expectedResponse{
				http.StatusBadRequest,
				"",
				expectedJSONContentType,
			},
		},
		{
			"Value missing metric type",
			http.MethodPost,
			"/value",
			"{}",
			"application/json",
			expectedResponse{
				http.StatusBadRequest,
				"",
				expectedJSONContentType,
			},
		},
		{
			"Value invalid metric type",
			http.MethodPost,
			"/value",
			"{\"type\":\"random\",\"id\":\"test\"}",
			"application/json",
			expectedResponse{
				http.StatusBadRequest,
				"",
				expectedJSONContentType,
			},
		},
		{
			"Value missing metric name",
			http.MethodPost,
			"/value",
			"{\"type\":\"gauge\"}",
			"application/json",
			expectedResponse{
				http.StatusBadRequest,
				"",
				expectedJSONContentType,
			},
		},

		// Missing or invalid metric VALUE
		{
			"Missing gauge metric value",
			http.MethodPost,
			"/update",
			"{\"type\":\"gauge\", \"id\":\"name\"}",
			"application/json",
			expectedResponse{
				http.StatusBadRequest,
				"",
				expectedJSONContentType,
			},
		},
		{
			"Missing counter metric value",
			http.MethodPost,
			"/update",
			"{\"type\":\"counter\", \"id\":\"name\"}",
			"application/json",
			expectedResponse{
				http.StatusBadRequest,
				"",
				expectedJSONContentType,
			},
		},
		{
			"Gauge metric value is non-numeric",
			http.MethodPost,
			"/update",
			"{\"type\":\"gauge\", \"id\":\"name\", \"value\":\"LatinText\"}",
			"application/json",
			expectedResponse{
				http.StatusBadRequest,
				"",
				expectedJSONContentType,
			},
		},
		{
			"Counter metric value is non-numeric",
			http.MethodPost,
			"/update",
			"{\"type\":\"counter\", \"id\":\"name\", \"delta\":\"LatinText\"}",
			"application/json",
			expectedResponse{
				http.StatusBadRequest,
				"",
				expectedJSONContentType,
			},
		},
		{
			"Counter metric value is positive float",
			http.MethodPost,
			"/update",
			"{\"type\":\"counter\", \"id\":\"name\", \"delta\":1001.00}",
			"application/json",
			expectedResponse{
				http.StatusBadRequest,
				"",
				expectedJSONContentType,
			},
		},
		{
			"Counter metric value is negative float",
			http.MethodPost,
			"/update",
			"{\"type\":\"counter\", \"id\":\"name\", \"delta\":-1001.00}",
			"application/json",
			expectedResponse{
				http.StatusBadRequest,
				"",
				expectedJSONContentType,
			},
		},
		{
			"Counter metric value is negative int",
			http.MethodPost,
			"/update",
			"{\"type\":\"counter\", \"id\":\"test\", \"delta\":-1000}",
			"application/json",
			expectedResponse{
				http.StatusBadRequest,
				"",
				expectedJSONContentType,
			},
		},

		// Invalid JSON syntax
		{
			"Invalid JSON syntax",
			http.MethodPost,
			"/update",
			`{invalid json`,
			"application/json",
			expectedResponse{
				http.StatusBadRequest,
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
