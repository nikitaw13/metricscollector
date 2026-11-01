package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nikitaw13/metricscollector/internal/mocks"
	"github.com/nikitaw13/metricscollector/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

var (
	errTestListAllMetrics = errors.New("failed to list all metrics")
	errTestGetGauge       = errors.New("failed to get a gauge")
	errTestGetCounter     = errors.New("failed to get a counter")
	errTestUpdateGauge    = errors.New("failed to update a gauge")
	errTestUpdateCounter  = errors.New("failed to update a counter")
	errTestUpdateBatch    = errors.New("failed to update a batch of metrics")
)

var (
	testMetricName   = "metric"
	mockGaugeValue   = float64(1.00)
	mockCounterValue = int64(1)
	testBatchBody    = fmt.Sprintf(`[{"type":"%s","id":"%s","delta":%d}]`, "counter", testMetricName, mockCounterValue)
	testBatchMetrics = []model.Metric{{ID: testMetricName, Type: "counter", Delta: &mockCounterValue}}
)

var (
	listAllMetricsEndpoint      = "/"
	updateJSONEndpoint          = "/update"
	readJSONEndpoint            = "/value"
	updateJSONBatchEndpoint     = "/updates"
	readGaugeMetricEndpoint     = fmt.Sprintf("/value/gauge/%s", testMetricName)
	readCounterMetricEndpoint   = fmt.Sprintf("/value/counter/%s", testMetricName)
	updateGaugeMetricEndpoint   = fmt.Sprintf("/update/gauge/%s/%g", testMetricName, mockGaugeValue)
	updateCounterMetricEndpoint = fmt.Sprintf("/update/counter/%s/%d", testMetricName, mockCounterValue)
)

func TestMetricsHandler_handleListMetricsError(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(svc *mocks.MockService)
		wantStatus  int
		wantMessage string
	}{
		{
			name: "service error returns 500",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().ListAllMetrics().Return(nil, nil, errTestListAllMetrics)
			},
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "Internal Server Error\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			svc := mocks.NewMockService(ctrl)
			tt.setup(svc)

			h := NewMetricsHandler("", svc, zap.NewNop())

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, listAllMetricsEndpoint, nil)

			h.Router().ServeHTTP(w, r)

			assert.Equal(t, tt.wantStatus, w.Result().StatusCode)
			assert.Equal(t, tt.wantMessage, w.Body.String())
		})
	}
}

func TestMetricsHandler_handleURLReadError(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(svc *mocks.MockService)
		endpoint    string
		wantStatus  int
		wantMessage string
	}{
		{
			name: "gauge service error returns 500",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().GetGauge(testMetricName).Return(mockGaugeValue, errTestGetGauge)
			},
			endpoint:    readGaugeMetricEndpoint,
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "Internal Server Error\n",
		},
		{
			name: "counter service error returns 500",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().GetCounter(testMetricName).Return(mockCounterValue, errTestGetCounter)
			},
			endpoint:    readCounterMetricEndpoint,
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "Internal Server Error\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			svc := mocks.NewMockService(ctrl)
			tt.setup(svc)

			h := NewMetricsHandler("", svc, zap.NewNop())

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, tt.endpoint, nil)

			h.Router().ServeHTTP(w, r)

			assert.Equal(t, tt.wantStatus, w.Result().StatusCode)
			assert.Equal(t, tt.wantMessage, w.Body.String())
		})
	}
}

func TestMetricsHandler_handleURLUpdateError(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(svc *mocks.MockService)
		endpoint    string
		wantStatus  int
		wantMessage string
	}{
		{
			name: "gauge service error returns 500",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().SetGauge(testMetricName, mockGaugeValue).Return(errTestUpdateGauge)
			},
			endpoint:    updateGaugeMetricEndpoint,
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "Internal Server Error\n",
		},
		{
			name: "counter service error returns 500",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().AddCounter(testMetricName, mockCounterValue).Return(mockCounterValue, errTestUpdateCounter)
			},
			endpoint:    updateCounterMetricEndpoint,
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "Internal Server Error\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			svc := mocks.NewMockService(ctrl)
			tt.setup(svc)

			h := NewMetricsHandler("", svc, zap.NewNop())

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, tt.endpoint, nil)

			h.Router().ServeHTTP(w, r)

			assert.Equal(t, tt.wantStatus, w.Result().StatusCode)
			assert.Equal(t, tt.wantMessage, w.Body.String())
		})
	}
}

func TestMetricsHandler_handleJSONUpdateError(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(svc *mocks.MockService)
		body        string
		wantStatus  int
		wantMessage string
	}{
		{
			name: "gauge service error returns 500",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().SetGauge(testMetricName, mockGaugeValue).Return(errTestUpdateGauge)
			},
			body:        fmt.Sprintf(`{"type":"%s","id":"%s","value":%g}`, "gauge", testMetricName, mockGaugeValue),
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "{\"code\":500,\"message\":\"Internal Server Error\"}",
		},
		{
			name: "counter service error returns 500",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().AddCounter(testMetricName, mockCounterValue).Return(mockCounterValue, errTestUpdateCounter)
			},
			body:        fmt.Sprintf(`{"type":"%s","id":"%s","delta":%d}`, "counter", testMetricName, mockCounterValue),
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "{\"code\":500,\"message\":\"Internal Server Error\"}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			svc := mocks.NewMockService(ctrl)
			tt.setup(svc)

			h := NewMetricsHandler("", svc, zap.NewNop())

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, updateJSONEndpoint, strings.NewReader(tt.body))
			r.Header.Set("Content-Type", "application/json")

			h.Router().ServeHTTP(w, r)

			assert.Equal(t, tt.wantStatus, w.Result().StatusCode)
			assert.Equal(t, tt.wantMessage, w.Body.String())
		})
	}
}

func TestMetricsHandler_handleJSONReadError(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(svc *mocks.MockService)
		body        string
		wantStatus  int
		wantMessage string
	}{
		{
			name: "gauge service error returns 500",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().GetGauge(testMetricName).Return(mockGaugeValue, errTestGetGauge)
			},
			body:        fmt.Sprintf(`{"type":"%s","id":"%s"}`, "gauge", testMetricName),
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "{\"code\":500,\"message\":\"Internal Server Error\"}",
		},
		{
			name: "counter service error returns 500",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().GetCounter(testMetricName).Return(mockCounterValue, errTestGetCounter)
			},
			body:        fmt.Sprintf(`{"type":"%s","id":"%s"}`, "counter", testMetricName),
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "{\"code\":500,\"message\":\"Internal Server Error\"}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			svc := mocks.NewMockService(ctrl)
			tt.setup(svc)

			h := NewMetricsHandler("", svc, zap.NewNop())

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, readJSONEndpoint, strings.NewReader(tt.body))
			r.Header.Set("Content-Type", "application/json")

			h.Router().ServeHTTP(w, r)

			assert.Equal(t, tt.wantStatus, w.Result().StatusCode)
			assert.Equal(t, tt.wantMessage, w.Body.String())
		})
	}
}

func TestMetricsHandler_handleJSONUpdatesError(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(svc *mocks.MockService)
		body        string
		wantStatus  int
		wantMessage string
	}{
		{
			name: "service error returns 500",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().UpdateMetricsBatch(gomock.Any(), testBatchMetrics).Return(errTestUpdateBatch)
			},
			body:        testBatchBody,
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "{\"code\":500,\"message\":\"Internal Server Error\"}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			svc := mocks.NewMockService(ctrl)
			tt.setup(svc)

			h := NewMetricsHandler("", svc, zap.NewNop())

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, updateJSONBatchEndpoint, strings.NewReader(tt.body))
			r.Header.Set("Content-Type", "application/json")

			h.Router().ServeHTTP(w, r)

			assert.Equal(t, tt.wantStatus, w.Result().StatusCode)
			assert.Equal(t, tt.wantMessage, w.Body.String())
		})
	}
}

// TestJSONClientErrorMessages verifies that 4xx responses carry the
// client-safe messages produced by decoding and validation — the
// counterpart of the 500 rule that hides internal error details.
func TestJSONClientErrorMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		path         string
		body         string
		wantContains string
	}{
		{
			name:         "value invalid JSON syntax carries decode message",
			path:         "/value",
			body:         `{"type": gauge}`,
			wantContains: "invalid JSON syntax at offset",
		},
		{
			name:         "value truncated body classified as empty",
			path:         "/value",
			body:         `{"type":"gauge", "id":`,
			wantContains: "request body is empty",
		},
		{
			name:         "update missing type carries validation message",
			path:         "/update",
			body:         `{}`,
			wantContains: "Metric type is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := newTestServer()
			t.Cleanup(ts.Close)

			resp := doJSONRequest(t, ts, http.MethodPost, tc.path, tc.body, "application/json")

			var got apiError
			require.NoError(t, json.Unmarshal([]byte(resp.body), &got))
			assert.Equal(t, http.StatusBadRequest, got.Code)
			assert.Contains(t, got.Message, tc.wantContains)
		})
	}
}
