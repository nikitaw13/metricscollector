package model

// Supported metric types.
const (
	Counter = "counter"
	Gauge   = "gauge"
)

// Metric represents a flat metric model.
// Delta and Value are pointers to distinguish 0 from an unset value.
// Type must be one of Counter or Gauge.
type Metric struct {
	ID    string   `json:"id"`
	Type  string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}
