package agent

import "testing"

func TestCollector_Run(t *testing.T) {
	type fields struct {
		Storage Storage
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Collector{
				Storage: tt.fields.Storage,
			}
			c.Run()
		})
	}
}
