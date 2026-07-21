package handler

import (
	"net/http"
	"testing"

	"github.com/PrometheRus/metricscollector/internal/repository"
)

func TestMetricsHandler_ServeHTTP(t *testing.T) {
	type fields struct {
		Storage repository.Repository
	}
	type args struct {
		res http.ResponseWriter
		req *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &MetricsHandler{
				Storage: tt.fields.Storage,
			}
			h.ServeHTTP(tt.args.res, tt.args.req)
		})
	}
}
