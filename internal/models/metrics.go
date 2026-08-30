// Package models содержит структуры данных и константы для работы с метриками.
package models

// Допустимые типы метрик.
const (
	Counter = "counter"
	Gauge   = "gauge"
)

// Metrics определяет параметры метрики.
type Metrics struct {
	// ID содержит название метрики.
	ID string `json:"id"`
	// MType определяет тип метрики.
	MType string `json:"type"`
	// Delta определяет значение метрики типа counter.
	Delta *int64 `json:"delta,omitempty"`
	// Value определяет значение метрики типа gauge.
	Value *float64 `json:"value,omitempty"`
}

// PointerFloat64 возвращает указатель на переданное значение float64.
func PointerFloat64(value float64) *float64 {
	return &value
}

// PointerInt64 возвращает указатель на переданное значение int64.
func PointerInt64(value int64) *int64 {
	return &value
}
