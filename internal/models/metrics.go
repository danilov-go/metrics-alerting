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

// Pointer возвращает указатель на переданное значение любого типа.
func Pointer[T any](value T) *T {
	p := new(T)
	*p = value
	return p
}
