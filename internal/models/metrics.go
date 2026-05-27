package models

const (
	Counter = "counter"
	Gauge   = "gauge"
)

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func PointerFloat64(value float64) *float64 {
	return &value
}

func PointerInt64(value int64) *int64 {
	return &value
}
