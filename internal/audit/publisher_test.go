package audit

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/nikitaw13/metricscollector/internal/model"
	"go.uber.org/mock/gomock"
)

var (
	testUnixTimestamp = time.Now().Unix()
	testMetricName    = "testMetricName"
	testAuditIP       = "127.0.0.1"

	testGaugeValue = float64(1.00)
	testMetric     = model.Metric{
		ID:    testMetricName,
		Type:  model.Gauge,
		Value: &testGaugeValue,
	}

	testJSONEvent = fmt.Appendf([]byte{}, `{"ts":%d,"metrics":[%q],"ip_address":%q}`, testUnixTimestamp, testMetricName, testAuditIP)

	testFilename = "testfilename.txt"
	testURL      = "http://localhost:8181"

	errSenderFailed       = errors.New("sender failed")
	errFirstSenderFailed  = errors.New("first sender failed")
	errSecondSenderFailed = errors.New("second sender failed")
	errThirdSenderFailed  = errors.New("third sender failed")
)

func TestNew(t *testing.T) {
	type args struct {
		file string
		url  string
	}
	tests := []struct {
		name string
		args args
		want *Auditor
	}{
		{
			name: "one file",
			args: args{file: testFilename},
			want: &Auditor{receivers: []subscriber{
				fileWriter{file: testFilename},
			}},
		},
		{
			name: "one url",
			args: args{url: testURL},
			want: &Auditor{receivers: []subscriber{
				urlSender{url: testURL},
			}},
		},
		{
			name: "both file & url",
			args: args{file: testFilename, url: testURL},
			want: &Auditor{receivers: []subscriber{
				fileWriter{file: testFilename},
				urlSender{url: testURL},
			}},
		},
		{
			name: "neither file nor url",
			args: args{},
			want: &Auditor{receivers: nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := New(tt.args.file, tt.args.url, nil)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
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
		receiverCount int
		setup         func(sender *Mocksubscriber)
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
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("Auditor.Notify() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && err != nil {
				t.Errorf("Auditor.Notify() unexpected error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuditor_notify(t *testing.T) {
	tests := []struct {
		name          string
		receiverCount int
		setup         func(sender *Mocksubscriber)
		event         []byte
		wantErr       []error
	}{
		{
			name:          "single receiver succeeds",
			receiverCount: 1,
			setup: func(sender *Mocksubscriber) {
				sender.EXPECT().send(testJSONEvent).Return(nil)
			},
			event:   testJSONEvent,
			wantErr: nil,
		},
		{
			name:          "triple receivers succeed",
			receiverCount: 3,
			setup: func(sender *Mocksubscriber) {
				sender.EXPECT().send(testJSONEvent).Times(3).Return(nil)
			},
			event:   testJSONEvent,
			wantErr: nil,
		},
		{
			name:          "single receiver fails",
			receiverCount: 1,
			setup: func(sender *Mocksubscriber) {
				sender.EXPECT().send(testJSONEvent).Return(errFirstSenderFailed)
			},
			event:   testJSONEvent,
			wantErr: []error{errFirstSenderFailed},
		},
		{
			name:          "triple receivers fail",
			receiverCount: 3,
			setup: func(sender *Mocksubscriber) {
				sender.EXPECT().send(testJSONEvent).Return(errFirstSenderFailed)
				sender.EXPECT().send(testJSONEvent).Return(errSecondSenderFailed)
				sender.EXPECT().send(testJSONEvent).Return(errThirdSenderFailed)
			},
			event:   testJSONEvent,
			wantErr: []error{errFirstSenderFailed, errSecondSenderFailed, errThirdSenderFailed},
		},
		{
			name:          "succeed, fail, succeed",
			receiverCount: 3,
			setup: func(sender *Mocksubscriber) {
				sender.EXPECT().send(testJSONEvent).Return(nil)
				sender.EXPECT().send(testJSONEvent).Return(errSecondSenderFailed)
				sender.EXPECT().send(testJSONEvent).Return(nil)
			},
			event:   testJSONEvent,
			wantErr: []error{errSecondSenderFailed},
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

			errs := auditor.notify(tt.event)

			if tt.wantErr == nil {
				if len(errs) != 0 {
					t.Errorf("notify() unexpected errors = %v", errs)
				}
				return
			}

			if len(errs) != len(tt.wantErr) {
				t.Fatalf("notify() errors = %v, want %v", errs, tt.wantErr)
			}
			for i, want := range tt.wantErr {
				if !errors.Is(errs[i], want) {
					t.Errorf("notify()[%d] = %v, want %v", i, errs[i], want)
				}
			}
		})
	}
}
