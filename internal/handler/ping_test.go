package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nikitaw13/metricscollector/internal/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

var errTestPingDatabase = errors.New("failed to ping database")

func TestMetricsHandler_handleDatabasePing(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(svc *mocks.MockService)
		wantStatus  int
		wantMessage string
	}{
		{
			name: "returns 200 and Pong when database is up",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().DatabasePing(gomock.Any()).Return(nil)
			},
			wantStatus:  http.StatusOK,
			wantMessage: "Pong\n",
		},
		{
			name: "database unavailable returns 500 with internal message",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().DatabasePing(gomock.Any()).Return(errTestPingDatabase)
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
			r := httptest.NewRequest(http.MethodGet, "/ping", nil)

			h.handleDatabasePing(w, r)
			assert.Equal(t, tt.wantStatus, w.Result().StatusCode)
			assert.Equal(t, tt.wantMessage, w.Body.String())
		})
	}
}
