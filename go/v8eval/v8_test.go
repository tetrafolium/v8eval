package v8eval

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	Initialize()
}

type pair struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func TestEval(t *testing.T) {
	v8 := NewV8()

	var i int
	assert.Equal(t, nil, v8.Eval("1 + 2", &i))
	assert.Equal(t, 3, i)

	var p pair
	assert.Equal(t, nil, v8.Eval("var p = { x: 1.1, y: 2.2 }; p", &p))
	assert.Equal(t, pair{X: 1.1, Y: 2.2}, p)

	assert.Equal(t, nil, v8.Eval("", nil))
	assert.Equal(t, nil, v8.Eval("function inc(x) { return x + 1; }", nil))

	err := v8.Eval("foo", nil)
	assert.NotNil(t, err)
	assert.Equal(t, "ReferenceError: foo is not defined", err.Error())

	var s string
	err = v8.Eval("1", &s)
	assert.NotNil(t, err)
	assert.Equal(t, "json: cannot unmarshal number into Go value of type string", err.Error())
}

func TestCall(t *testing.T) {
	v8 := NewV8()
	v8.Eval("function inc(x) { return x + 1; }", nil)

	var i int
	assert.Equal(t, nil, v8.Call("inc", []int{7}, &i))
	assert.Equal(t, 8, i)

	err := v8.Call("i", []int{7}, &i)
	assert.NotNil(t, err)
	assert.Equal(t, "TypeError: 'i' is not a function", err.Error())

	err = v8.Call("inc", nil, &i)
	assert.NotNil(t, err)
	assert.Equal(t, "TypeError: 'null' is not an array", err.Error())

	var s string
	err = v8.Call("inc", []int{7}, &s)
	assert.NotNil(t, err)
	assert.Equal(t, "json: cannot unmarshal number into Go value of type string", err.Error())
}

func TestInParallel(t *testing.T) {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)

	ch := make(chan int)

	loop := func(n int) {
		v8 := NewV8()
		v8.Eval("function inc(x) { return x + 1; }", nil)
		i := 0
		for i < n {
			v8.Call("inc", []int{i}, &i)
		}
		ch <- i
	}

	const numRepeat = 1000
	const numGoroutine = 10
	for i := 0; i < numGoroutine; i++ {
		go loop(numRepeat)
	}

	for i := 0; i < numGoroutine; i++ {
		x := <-ch
		assert.Equal(t, numRepeat, x)
	}
}
