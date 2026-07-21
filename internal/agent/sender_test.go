Generated TestSender_Run
package agent

import "testing"

func TestSender_Run(t *testing.T) {
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
			s := &Sender{
				Storage: tt.fields.Storage,
			}
			s.Run()
		})
	}
}
