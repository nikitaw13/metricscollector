package retry

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

var errSimulatedReset = errors.New("simulated connection reset")

func TestClientWithRetries_Do(t *testing.T) {
	type args struct {
		errors      []error
		status      int
		retryDelays []time.Duration
		rewindable  bool
	}
	tests := []struct {
		name       string
		args       args
		wantStatus int
		wantErr    error
		wantCalls  int
	}{
		{
			name: "succeeds after two failures",
			args: args{
				errors:      []error{errSimulatedReset, errSimulatedReset},
				retryDelays: []time.Duration{time.Millisecond, time.Millisecond},
				rewindable:  true,
			},
			wantStatus: http.StatusOK,
			wantCalls:  3,
		},
		{
			name: "retries exhausted",
			args: args{
				errors:      []error{errSimulatedReset, errSimulatedReset, errSimulatedReset, errSimulatedReset},
				retryDelays: []time.Duration{time.Millisecond, time.Millisecond, time.Millisecond},
				rewindable:  true,
			},
			wantErr:   errSimulatedReset,
			wantCalls: 4,
		},
		{
			name: "succeeds on first attempt",
			args: args{
				retryDelays: []time.Duration{time.Millisecond, time.Millisecond, time.Millisecond},
				rewindable:  true,
			},
			wantStatus: http.StatusOK,
			wantCalls:  1,
		},
		{
			name: "no retry on non rewindable body",
			args: args{
				errors:      []error{errSimulatedReset},
				retryDelays: []time.Duration{time.Millisecond, time.Millisecond, time.Millisecond},
				rewindable:  false,
			},
			wantErr:   errSimulatedReset,
			wantCalls: 1,
		},
		{
			name: "no retry on server error response",
			args: args{
				status:      http.StatusInternalServerError,
				retryDelays: []time.Duration{time.Millisecond, time.Millisecond, time.Millisecond},
				rewindable:  true,
			},
			wantStatus: http.StatusInternalServerError,
			wantCalls:  1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fake := &fakeHTTPClient{errors: tt.args.errors, status: tt.args.status}
			client := New(tt.args.retryDelays, fake, zap.NewNop())

			req := newRetryRequest(t)

			if !tt.args.rewindable {
				// The anonymous struct hides *strings.Reader from http.NewRequest,
				// so the request is built without a GetBody function.
				var err error
				req, err = http.NewRequest(
					http.MethodPost,
					"http://retry.test/updates",
					struct{ io.Reader }{strings.NewReader("{}")},
				)
				if err != nil {
					t.Fatalf("error building request = %v", err)
				}
			}

			statusCode, err := doRequest(t, client, req)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ClientWithRetries.Do() error = %v, wantErr %v", err, tt.wantErr)
			}
			// wantStatus is 0 for error cases: doRequest returns 0 on failure.
			if statusCode != tt.wantStatus {
				t.Errorf("ClientWithRetries.Do() status = %d, want %d", statusCode, tt.wantStatus)
			}
			if fake.calls != tt.wantCalls {
				t.Errorf("ClientWithRetries.Do() calls = %d, want %d", fake.calls, tt.wantCalls)
			}
		})
	}
}
