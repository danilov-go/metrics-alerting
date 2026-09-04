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
	poolTest := New[testStuctur]()
	test := poolTest.Get()
	assert.NotNil(t, test)
	assert.Equal(t, 0, test.value)
	assert.Equal(t, 0.0, test.float)
	assert.Equal(t, "", test.str)
	test.value = 10
	test.float = 123.45
	test.str = "test"
	poolTest.Put(test)
	testZero := poolTest.Get()
	assert.NotNil(t, testZero)
	assert.Equal(t, 0, testZero.value)
	assert.Equal(t, 0.0, testZero.float)
	assert.Equal(t, "", testZero.str)
}
