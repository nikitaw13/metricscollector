package repository

import (
	"reflect"
	"testing"
)

func TestNewMemStorage(t *testing.T) {
	tests := []struct {
		name string
		want *MemStorage
	}{
		{
			name: "Return NewMemStorage instance",
			want: &MemStorage{
				gauge:   map[string]float64{},
				counter: map[string]int64{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMemStorage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMemStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemStorage_UpdateGauge(t *testing.T) {
	type fields struct {
		gauge   map[string]float64
		counter map[string]int64
	}
	type args struct {
		name  string
		value float64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test metric_gauge is 1.00",
			fields: fields{
				gauge:   map[string]float64{"metric_gauge": 1.00},
				counter: map[string]int64{"metric_counter": 1},
			},
			args:    args{name: "metric_gauge", value: 1.00},
			wantErr: false,
		},
		{
			name: "Test metric_gauge is -1.00",
			fields: fields{
				gauge:   map[string]float64{"metric_gauge": -1.00},
				counter: map[string]int64{"metric_counter": 1},
			},
			args:    args{name: "metric_gauge", value: -1.00},
			wantErr: false,
		},
		{
			name: "Test metric_gauge is 1111.50",
			fields: fields{
				gauge:   map[string]float64{"metric_gauge": 1111.50},
				counter: map[string]int64{"metric_counter": 1},
			},
			args:    args{name: "metric_gauge", value: 1111.50},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				gauge:   tt.fields.gauge,
				counter: tt.fields.counter,
			}
			if err := ms.UpdateGauge(tt.args.name, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("MemStorage.UpdateGauge() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemStorage_UpdateCounter(t *testing.T) {
	type fields struct {
		gauge   map[string]float64
		counter map[string]int64
	}
	type args struct {
		name  string
		value int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test metric_counter is 1",
			fields: fields{
				gauge:   map[string]float64{"metric_gauge": 1.00},
				counter: map[string]int64{"metric_counter": 1},
			},
			args:    args{name: "metric_counter", value: 1},
			wantErr: false,
		},
		{
			name: "Test metric_counter is -11111",
			fields: fields{
				gauge:   map[string]float64{"metric_gauge": 1.00},
				counter: map[string]int64{"metric_counter": -1111},
			},
			args:    args{name: "metric_counter", value: -1111},
			wantErr: false,
		},
		{
			name: "Test metric_counter is 1111",
			fields: fields{
				gauge:   map[string]float64{"metric_gauge": 1.00},
				counter: map[string]int64{"metric_counter": 1},
			},
			args:    args{name: "metric_counter", value: 1111},
			wantErr: false,
		},
		{
			name: "Test metric_counter is 0",
			fields: fields{
				gauge:   map[string]float64{"metric_gauge": 1.00},
				counter: map[string]int64{"metric_counter": 1},
			},
			args:    args{name: "metric_counter", value: 0},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				gauge:   tt.fields.gauge,
				counter: tt.fields.counter,
			}
			if err := ms.UpdateCounter(tt.args.name, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("MemStorage.UpdateCounter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
