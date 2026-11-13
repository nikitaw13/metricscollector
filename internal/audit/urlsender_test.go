package audit

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_urlSender_send(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			w.WriteHeader(http.StatusOK)
		case "/redirect":
			w.WriteHeader(http.StatusTemporaryRedirect)
		case "/clienterror":
			w.WriteHeader(http.StatusBadRequest)
		case "/servererror":
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	t.Cleanup(ts.Close)

	tests := []struct {
		name    string
		path    string
		event   []byte
		wantErr error
	}{
		{
			name:    "ok",
			path:    testPathOK,
			event:   testEvent,
			wantErr: nil,
		},
		{
			name:    "redirect",
			path:    testPathRedirect,
			event:   testEvent,
			wantErr: fmt.Errorf("unexpected redirect: status code %d", http.StatusTemporaryRedirect),
		},
		{
			name:    "client error",
			path:    testPathClientError,
			event:   testEvent,
			wantErr: fmt.Errorf("client error while sending event: status code %d", http.StatusBadRequest),
		},
		{
			name:    "server error",
			path:    testPathServerError,
			event:   testEvent,
			wantErr: fmt.Errorf("server error while sending event: status code %d", http.StatusInternalServerError),
		},
		{
			// unmatched path: the handler writes nothing, so the server
			// responds with an implicit 200 and send reports no error.
			name:    "implicit ok",
			path:    testPathUnmatched,
			event:   testEvent,
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			us := urlSender{
				url:    ts.URL + tt.path,
				client: http.DefaultClient,
			}

			err := us.send(tt.event)

			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
