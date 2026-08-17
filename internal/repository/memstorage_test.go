package repository

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
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
			if got := New(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetGauge(t *testing.T) {
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
			name: "gauge value is 1.00",
			fields: fields{
				gauge:   map[string]float64{"metric_gauge": 1.00},
				counter: map[string]int64{"metric_counter": 1},
			},
			args:    args{name: "metric_gauge", value: 1.00},
			wantErr: false,
		},
		{
			name: "gauge value is -1.00",
			fields: fields{
				gauge:   map[string]float64{"metric_gauge": -1.00},
				counter: map[string]int64{"metric_counter": 1},
			},
			args:    args{name: "metric_gauge", value: -1.00},
			wantErr: false,
		},
		{
			name: "gauge value is 1111.50",
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
			if err := ms.SetGauge(tt.args.name, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("MemStorage.SetGauge() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAddCounter(t *testing.T) {
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
			name: "counter value is 1",
			fields: fields{
				gauge:   map[string]float64{"metric_gauge": 1.00},
				counter: map[string]int64{"metric_counter": 1},
			},
			args:    args{name: "metric_counter", value: 1},
			wantErr: false,
		},
		{
			name: "counter value is -11111",
			fields: fields{
				gauge:   map[string]float64{"metric_gauge": 1.00},
				counter: map[string]int64{"metric_counter": -1111},
			},
			args:    args{name: "metric_counter", value: -1111},
			wantErr: false,
		},
		{
			name: "counter value is 1111",
			fields: fields{
				gauge:   map[string]float64{"metric_gauge": 1.00},
				counter: map[string]int64{"metric_counter": 1},
			},
			args:    args{name: "metric_counter", value: 1111},
			wantErr: false,
		},
		{
			name: "counter value is 0",
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
			if err := ms.AddCounter(tt.args.name, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("MemStorage.AddCounter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
