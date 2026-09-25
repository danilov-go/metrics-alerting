package models_test

import (
	"testing"

	"github.com/danilov-go/metrics-alerting.git/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestPointer(t *testing.T) {
	float := 123.45
	num := 5
	str := "test"
	boolean := true
	type want struct {
		val    any
		valExp any
	}
	tests := []struct {
		name string
		exp  want
	}{
		{
			name: "Указатель на int",
			exp: want{
				val:    num,
				valExp: &num,
			},
		},
		{
			name: "Указатель на float64",
			exp: want{
				val:    float,
				valExp: &float,
			},
		},
		{
			name: "Указатель на string",
			exp: want{
				val:    str,
				valExp: &str,
			},
		},
		{
			name: "Указатель на bool",
			exp: want{
				val:    boolean,
				valExp: &boolean,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result any
			switch v := tt.exp.val.(type) {
			case int:
				result = models.Pointer(v)
			case float64:
				result = models.Pointer(v)
			case string:
				result = models.Pointer(v)
			case bool:
				result = models.Pointer(v)
			default:
				t.Fatalf("неизвестный тип данных: %T", v)
			}
			assert.Equal(t, tt.exp.valExp, result)
		})
	}
}
