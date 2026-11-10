package audit

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

var (
	testNonExistingPath      = "ABRACADABRA/testfilename.txt"
	testEvent                = []byte("first event")
	testSecondEvent          = []byte("second event")
	testComparingEvent       = []byte("first event\n")
	testComparingDoubleEvent = []byte("first event\nsecond event\n")
)

var (
	testPathOK          = "/ok"
	testPathRedirect    = "/redirect"
	testPathClientError = "/clienterror"
	testPathServerError = "/servererror"
	testPathUnmatched   = "/unmatched"
)

func Test_fileWriter_send(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		filename      string
		seededContent []byte
		event         []byte
		wantEvent     []byte
		wantErr       error
	}{
		{
			name:          "one event to write",
			filename:      testFilename,
			seededContent: nil,
			event:         testEvent,
			wantEvent:     testComparingEvent,
			wantErr:       nil,
		},
		{
			name:          "two events to write",
			filename:      testFilename,
			seededContent: testEvent,
			event:         testSecondEvent,
			wantEvent:     testComparingDoubleEvent,
			wantErr:       nil,
		},
		{
			name:          "non existing directory",
			filename:      testNonExistingPath,
			seededContent: nil,
			event:         testEvent,
			wantEvent:     nil,
			wantErr:       os.ErrNotExist,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			file := filepath.Join(t.TempDir(), tt.filename)
			fw := fileWriter{
				file: file,
			}

			if tt.seededContent != nil {
				openedFile, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, fileMode)
				if err != nil {
					t.Fatalf("error opening test file = %v", err)
				}
				defer openedFile.Close()

				_, err = fmt.Fprintf(openedFile, "%s\n", tt.seededContent)
				if err != nil {
					t.Fatalf("error pre-population test file = %v", err)
				}
			}

			err := fw.send(tt.event)

			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("fileWriter.send() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && err != nil {
				t.Errorf("fileWriter.send() unexpected error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr == nil {
				gotContent, err := os.ReadFile(file)
				if err != nil {
					t.Fatalf("error reading file = %v", err)
				}
				if !bytes.Equal(gotContent, tt.wantEvent) {
					t.Errorf("events do not equal: got %q, want %q", gotContent, tt.wantEvent)
				}
			}

		})
	}
}

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
			name:    "unmatched path",
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

			if tt.wantErr != nil && (err == nil || err.Error() != tt.wantErr.Error()) {
				t.Errorf("urlSender.send() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && err != nil {
				t.Errorf("urlSender.send() unexpected error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
