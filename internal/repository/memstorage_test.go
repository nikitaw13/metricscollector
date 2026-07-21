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
		// TODO: Add test cases.
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
		// TODO: Add test cases.
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
		// TODO: Add test cases.
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
