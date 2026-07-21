Generated TestNewAgentStorage
Generated TestAgentStorage_UpdateGauge
Generated TestAgentStorage_UpdateCounter
Generated TestAgentStorage_GetGauge
Generated TestAgentStorage_GetCounter
Generated TestAgentStorage_GetAllGauges
Generated TestAgentStorage_GetAllCounters
package agent

import (
	"reflect"
	"testing"
)

func TestNewAgentStorage(t *testing.T) {
	tests := []struct {
		name string
		want *AgentStorage
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAgentStorage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAgentStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgentStorage_UpdateGauge(t *testing.T) {
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
			ms := &AgentStorage{
				gauge:   tt.fields.gauge,
				counter: tt.fields.counter,
			}
			if err := ms.UpdateGauge(tt.args.name, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("AgentStorage.UpdateGauge() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAgentStorage_UpdateCounter(t *testing.T) {
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
			ms := &AgentStorage{
				gauge:   tt.fields.gauge,
				counter: tt.fields.counter,
			}
			if err := ms.UpdateCounter(tt.args.name, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("AgentStorage.UpdateCounter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAgentStorage_GetGauge(t *testing.T) {
	type fields struct {
		gauge   map[string]float64
		counter map[string]int64
	}
	type args struct {
		name string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantValue float64
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &AgentStorage{
				gauge:   tt.fields.gauge,
				counter: tt.fields.counter,
			}
			gotValue, err := ms.GetGauge(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Fatalf("AgentStorage.GetGauge() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if gotValue != tt.wantValue {
				t.Errorf("AgentStorage.GetGauge() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func TestAgentStorage_GetCounter(t *testing.T) {
	type fields struct {
		gauge   map[string]float64
		counter map[string]int64
	}
	type args struct {
		name string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantValue int64
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &AgentStorage{
				gauge:   tt.fields.gauge,
				counter: tt.fields.counter,
			}
			gotValue, err := ms.GetCounter(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Fatalf("AgentStorage.GetCounter() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if gotValue != tt.wantValue {
				t.Errorf("AgentStorage.GetCounter() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func TestAgentStorage_GetAllGauges(t *testing.T) {
	type fields struct {
		gauge   map[string]float64
		counter map[string]int64
	}
	tests := []struct {
		name   string
		fields fields
		wantM  map[string]float64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &AgentStorage{
				gauge:   tt.fields.gauge,
				counter: tt.fields.counter,
			}
			if gotM := ms.GetAllGauges(); !reflect.DeepEqual(gotM, tt.wantM) {
				t.Errorf("AgentStorage.GetAllGauges() = %v, want %v", gotM, tt.wantM)
			}
		})
	}
}

func TestAgentStorage_GetAllCounters(t *testing.T) {
	type fields struct {
		gauge   map[string]float64
		counter map[string]int64
	}
	tests := []struct {
		name   string
		fields fields
		wantM  map[string]int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &AgentStorage{
				gauge:   tt.fields.gauge,
				counter: tt.fields.counter,
			}
			if gotM := ms.GetAllCounters(); !reflect.DeepEqual(gotM, tt.wantM) {
				t.Errorf("AgentStorage.GetAllCounters() = %v, want %v", gotM, tt.wantM)
			}
		})
	}
}
