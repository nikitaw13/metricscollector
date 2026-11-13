// Package audit publishes metric-update events to the configured receivers.
package audit

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/nikitaw13/metricscollector/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

func TestAuditor_Close(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(sender *Mocksubscriber)
		receiverCount int
		wantErr       error
	}{
		{
			name: "all receivers close cleanly",
			setup: func(sender *Mocksubscriber) {
				sender.EXPECT().close().Times(3).Return(nil)
			},
			receiverCount: 3,
			wantErr:       nil,
		},
		{
			name: "all three receivers fail to close",
			setup: func(sender *Mocksubscriber) {
				sender.EXPECT().close().Times(3).Return(errCloseFailed)
			},
			receiverCount: 3,
			wantErr:       errCloseFailed,
		},
		{
			name: "one of three receivers fails to close",
			setup: func(sender *Mocksubscriber) {
				sender.EXPECT().close().Return(nil)
				sender.EXPECT().close().Return(errCloseFailed)
				sender.EXPECT().close().Return(nil)
			},
			receiverCount: 3,
			wantErr:       errCloseFailed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			sender := NewMocksubscriber(ctrl)
			tt.setup(sender)

			receivers := make([]subscriber, tt.receiverCount)
			for i := range receivers {
				receivers[i] = sender
			}

			auditor := &Auditor{receivers: receivers}

			err := auditor.Close()
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		fileName string
		url      string
		client   httpClient
	}
	tests := []struct {
		name      string
		args      args
		wantFiles int
		wantURLs  int
		wantErr   error
	}{
		{
			name:      "one file",
			args:      args{fileName: testFilename},
			wantFiles: 1,
			wantErr:   nil,
		},
		{
			name:      "non existing directory",
			args:      args{fileName: testNonExistingPath},
			wantFiles: 0,
			wantErr:   os.ErrNotExist,
		},
		{
			name:     "one url",
			args:     args{url: testURL, client: http.DefaultClient},
			wantURLs: 1,
			wantErr:  nil,
		},
		{
			name:      "both file and url",
			args:      args{fileName: testFilename, url: testURL, client: http.DefaultClient},
			wantFiles: 1,
			wantURLs:  1,
			wantErr:   nil,
		},
		{
			name:    "neither file nor url",
			args:    args{},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			filePath := tt.args.fileName
			if filePath != "" {
				filePath = filepath.Join(t.TempDir(), filePath)
			}

			got, err := New(filePath, tt.args.url, tt.args.client)
			require.ErrorIs(t, err, tt.wantErr)
			if err == nil {
				defer got.Close()
			}

			var files, urls int
			for _, receiver := range got.receivers {
				switch receiver.(type) {
				case *fileWriter:
					files++
				case *urlSender:
					urls++
				}
			}
			assert.Equal(t, tt.wantFiles, files)
			assert.Equal(t, tt.wantURLs, urls)
		})
	}
}

func TestAuditor_Notify(t *testing.T) {
	type args struct {
		timestamp int64
		metrics   []model.Metric
		ipAddress string
	}
	tests := []struct {
		name          string
		setup         func(sender *Mocksubscriber)
		receiverCount int
		args          args
		wantErr       error
	}{
		{
			name:          "single receiver succeeds",
			receiverCount: 1,
			setup: func(sender *Mocksubscriber) {
				sender.EXPECT().send(testJSONEvent).Return(nil)
			},
			args: args{
				timestamp: testUnixTimestamp,
				metrics:   []model.Metric{testMetric},
				ipAddress: testAuditIP,
			},
			wantErr: nil,
		},
		{
			name:          "three receivers succeed",
			receiverCount: 3,
			setup: func(sender *Mocksubscriber) {
				sender.EXPECT().send(testJSONEvent).Times(3).Return(nil)
			},
			args: args{
				timestamp: testUnixTimestamp,
				metrics:   []model.Metric{testMetric},
				ipAddress: testAuditIP,
			},
			wantErr: nil,
		},
		{
			name:          "single receiver fails",
			receiverCount: 1,
			setup: func(sender *Mocksubscriber) {
				sender.EXPECT().send(testJSONEvent).Return(errSenderFailed)
			},
			args: args{
				timestamp: testUnixTimestamp,
				metrics:   []model.Metric{testMetric},
				ipAddress: testAuditIP,
			},
			wantErr: errSenderFailed,
		},
		{
			name:          "three receivers fail",
			receiverCount: 3,
			setup: func(sender *Mocksubscriber) {
				sender.EXPECT().send(testJSONEvent).Times(3).Return(errSenderFailed)
			},
			args: args{
				timestamp: testUnixTimestamp,
				metrics:   []model.Metric{testMetric},
				ipAddress: testAuditIP,
			},
			wantErr: errSenderFailed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			sender := NewMocksubscriber(ctrl)
			tt.setup(sender)

			receivers := make([]subscriber, tt.receiverCount)
			for i := range receivers {
				receivers[i] = sender
			}

			auditor := &Auditor{receivers: receivers}

			err := auditor.Notify(tt.args.timestamp, tt.args.metrics, tt.args.ipAddress)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestAuditor_notify(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(sender *Mocksubscriber)
		receiverCount int
		event         []byte
		wantErrCount  int
	}{
		{
			name: "no receivers, no errors",
			setup: func(sender *Mocksubscriber) {
			},
			receiverCount: 0,
			event:         testEvent,
			wantErrCount:  0,
		},
		{
			name: "three receivers, no errors",
			setup: func(sender *Mocksubscriber) {
				sender.EXPECT().send(testEvent).Times(3).Return(nil)
			},
			receiverCount: 3,
			event:         testEvent,
			wantErrCount:  0,
		},
		{
			name: "three receivers, one error",
			setup: func(sender *Mocksubscriber) {
				sender.EXPECT().send(testEvent).Return(nil)
				sender.EXPECT().send(testEvent).Return(errSenderFailed)
				sender.EXPECT().send(testEvent).Return(nil)
			},
			receiverCount: 3,
			event:         testEvent,
			wantErrCount:  1,
		},
		{
			name: "three receivers, three errors",
			setup: func(sender *Mocksubscriber) {
				sender.EXPECT().send(testEvent).Return(errSenderFailed)
				sender.EXPECT().send(testEvent).Return(errSenderFailed)
				sender.EXPECT().send(testEvent).Return(errSenderFailed)
			},
			receiverCount: 3,
			event:         testEvent,
			wantErrCount:  3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			sender := NewMocksubscriber(ctrl)
			tt.setup(sender)

			receivers := make([]subscriber, tt.receiverCount)
			for i := range receivers {
				receivers[i] = sender
			}
			auditor := &Auditor{
				receivers: receivers,
			}

			errs := auditor.notify(tt.event)
			var errCount int
			for i := range errs {
				if errs[i] != nil {
					errCount++
				}
			}

			require.Equal(t, tt.wantErrCount, errCount)
		})
	}

}

func TestAuditor_subscribe(t *testing.T) {
	tests := []struct {
		name          string
		receiverCount int
		wantLength    int
	}{
		{
			name:          "several receivers",
			receiverCount: 5,
			wantLength:    5,
		},
		{
			name:          "no receivers",
			receiverCount: 0,
			wantLength:    0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			sender := NewMocksubscriber(ctrl)

			auditor := &Auditor{}
			for i := 0; i < tt.receiverCount; i++ {
				auditor.subscribe(sender)
			}
			assert.Equal(t, tt.wantLength, len(auditor.receivers))

			// subscribe one more unique receiver
			uniqueSender := NewMocksubscriber(ctrl)
			auditor.subscribe(uniqueSender)
			assert.Equal(t, tt.wantLength+1, len(auditor.receivers))
			assert.Contains(t, auditor.receivers, uniqueSender)

			// and one more
			anotherUniqueSender := NewMocksubscriber(ctrl)
			auditor.subscribe(anotherUniqueSender)
			assert.Equal(t, tt.wantLength+2, len(auditor.receivers))
			assert.Contains(t, auditor.receivers, anotherUniqueSender)
		})
	}
}

func TestAuditor_unsubscribe(t *testing.T) {
	tests := []struct {
		name          string
		receiverCount int
		wantLength    int
	}{
		{
			name:          "unsubscribe from five receivers",
			receiverCount: 5,
			wantLength:    5,
		},
		{
			name:          "unsubscribe from empty auditor",
			receiverCount: 0,
			wantLength:    0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			sender := NewMocksubscriber(ctrl)

			auditor := &Auditor{}
			for i := 0; i < tt.receiverCount; i++ {
				auditor.subscribe(sender)
			}
			assert.Equal(t, tt.wantLength, len(auditor.receivers))

			uniqueSender := NewMocksubscriber(ctrl)

			auditor.subscribe(uniqueSender)
			auditor.unsubscribe(uniqueSender)

			assert.Equal(t, tt.wantLength, len(auditor.receivers))

			for _, receiver := range auditor.receivers {
				assert.NotSame(t, receiver, uniqueSender)
			}

			anotherUniqueSender := NewMocksubscriber(ctrl)
			auditor.unsubscribe(anotherUniqueSender)
			assert.Equal(t, tt.wantLength, len(auditor.receivers))
		})
	}
}
