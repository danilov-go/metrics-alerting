package pool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStuctur struct {
	value int
	float float64
	str   string
}

func (rs *testStuctur) Reset() {
	if rs == nil {
		return
	}
	rs.value = 0
	rs.float = 0
	rs.str = ""
}

func TestPool(t *testing.T) {
	zeroStructur := testStuctur{
		value: 0,
		float: 0,
		str:   "",
	}
	tests := []struct {
		name         string
		factory      func() *testStuctur
		wantStructur testStuctur
	}{
		{
			name:         "передача nil функции-инициализатора",
			factory:      nil,
			wantStructur: zeroStructur,
		},
		{
			name: "передача заданной функции-инициализатора",
			factory: func() *testStuctur {
				return &testStuctur{
					value: 10,
					float: 123.45,
					str:   "test",
				}
			},
			wantStructur: testStuctur{
				value: 10,
				float: 123.45,
				str:   "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolTest := New[testStuctur](tt.factory)
			var test *testStuctur
			assert.NotPanics(t, func() {
				test = poolTest.Get()
			})
			assert.NotNil(t, test)
			assert.Equal(t, tt.wantStructur, *test)
			test.value = 15
			test.float = 223.45
			test.str = "testPut"
			poolTest.Put(test)
			testZero := poolTest.Get()
			assert.NotNil(t, testZero)
			assert.Equal(t, zeroStructur, *testZero)
		})
	}
}
