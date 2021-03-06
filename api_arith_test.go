package tensor

import (
	"log"
	"math/rand"
	"testing"
	"testing/quick"
	"time"

	"github.com/stretchr/testify/assert"
)

// This file contains the tests for API functions that aren't generated by genlib

func TestMod(t *testing.T) {
	a := New(WithBacking([]float64{1, 2, 3, 4}))
	b := New(WithBacking([]float64{1, 1, 1, 1}))
	var correct interface{} = []float64{0, 0, 0, 0}

	// vec-vec
	res, err := Mod(a, b)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())

	// scalar
	if res, err = Mod(a, 1.0); err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())
}

func TestFMA(t *testing.T) {
	same := func(q *Dense) bool {
		a := q.Clone().(*Dense)
		x := q.Clone().(*Dense)
		y := New(Of(q.Dtype()), WithShape(q.Shape().Clone()...))
		y.Memset(identityVal(100, q.Dtype()))
		WithEngine(q.Engine())(y)
		y2 := y.Clone().(*Dense)

		we, willFailEq := willerr(a, numberTypes, nil)
		_, ok1 := q.Engine().(FMAer)
		_, ok2 := q.Engine().(Muler)
		_, ok3 := q.Engine().(Adder)
		we = we || (!ok1 && (!ok2 || !ok3))

		f, err := FMA(a, x, y)
		if err, retEarly := qcErrCheck(t, "FMA#1", a, x, we, err); retEarly {
			if err != nil {
				log.Printf("q.Engine() %T", q.Engine())
				return false
			}
			return true
		}

		we, _ = willerr(a, numberTypes, nil)
		_, ok := a.Engine().(Muler)
		we = we || !ok
		wi, err := Mul(a, x, WithIncr(y2))
		if err, retEarly := qcErrCheck(t, "FMA#2", a, x, we, err); retEarly {
			if err != nil {
				return false
			}
			return true
		}
		return qcEqCheck(t, q.Dtype(), willFailEq, wi, f)
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if err := quick.Check(same, &quick.Config{Rand: r}); err != nil {
		t.Error(err)
	}

	// specific engines
	var eng Engine

	// FLOAT64 ENGINE

	// vec-vec
	eng = Float64Engine{}
	a := New(WithBacking(Range(Float64, 0, 100)), WithEngine(eng))
	x := New(WithBacking(Range(Float64, 1, 101)), WithEngine(eng))
	y := New(Of(Float64), WithShape(100), WithEngine(eng))

	f, err := FMA(a, x, y)
	if err != nil {
		t.Fatal(err)
	}

	a2 := New(WithBacking(Range(Float64, 0, 100)))
	x2 := New(WithBacking(Range(Float64, 1, 101)))
	y2 := New(Of(Float64), WithShape(100))
	f2, err := Mul(a2, x2, WithIncr(y2))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, f.Data(), f2.Data())

	// vec-scalar
	a = New(WithBacking(Range(Float64, 0, 100)), WithEngine(eng))
	y = New(Of(Float64), WithShape(100))

	if f, err = FMA(a, 2.0, y); err != nil {
		t.Fatal(err)
	}

	a2 = New(WithBacking(Range(Float64, 0, 100)))
	y2 = New(Of(Float64), WithShape(100))
	if f2, err = Mul(a2, 2.0, WithIncr(y2)); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, f.Data(), f2.Data())

	// FLOAT32 engine
	eng = Float32Engine{}
	a = New(WithBacking(Range(Float32, 0, 100)), WithEngine(eng))
	x = New(WithBacking(Range(Float32, 1, 101)), WithEngine(eng))
	y = New(Of(Float32), WithShape(100), WithEngine(eng))

	f, err = FMA(a, x, y)
	if err != nil {
		t.Fatal(err)
	}

	a2 = New(WithBacking(Range(Float32, 0, 100)))
	x2 = New(WithBacking(Range(Float32, 1, 101)))
	y2 = New(Of(Float32), WithShape(100))
	f2, err = Mul(a2, x2, WithIncr(y2))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, f.Data(), f2.Data())

	// vec-scalar
	a = New(WithBacking(Range(Float32, 0, 100)), WithEngine(eng))
	y = New(Of(Float32), WithShape(100))

	if f, err = FMA(a, float32(2), y); err != nil {
		t.Fatal(err)
	}

	a2 = New(WithBacking(Range(Float32, 0, 100)))
	y2 = New(Of(Float32), WithShape(100))
	if f2, err = Mul(a2, float32(2), WithIncr(y2)); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, f.Data(), f2.Data())

}

func TestMulScalarScalar(t *testing.T) {
	// scalar-scalar
	a := New(WithBacking([]float64{2}))
	b := New(WithBacking([]float64{3}))
	var correct interface{} = 6.0

	res, err := Mul(a, b)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())

	// Test commutativity
	res, err = Mul(b, a)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())

	// scalar-tensor
	a = New(WithBacking([]float64{3, 2}))
	b = New(WithBacking([]float64{2}))
	correct = []float64{6, 4}

	res, err = Mul(a, b)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())

	// Test commutativity
	res, err = Mul(b, a)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())

	// tensor - tensor
	a = New(WithBacking([]float64{3, 5}))
	b = New(WithBacking([]float64{7, 2}))
	correct = []float64{21, 10}

	res, err = Mul(a, b)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())

	// Test commutativity
	res, err = Mul(b, a)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())
}

func TestDivScalarScalar(t *testing.T) {
	// scalar-scalar
	a := New(WithBacking([]float64{6}))
	b := New(WithBacking([]float64{2}))
	var correct interface{} = 3.0

	res, err := Div(a, b)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())

	// scalar-tensor
	a = New(WithBacking([]float64{6, 4}))
	b = New(WithBacking([]float64{2}))
	correct = []float64{3, 2}

	res, err = Div(a, b)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())

	// tensor-scalar
	a = New(WithBacking([]float64{6}))
	b = New(WithBacking([]float64{3, 2}))
	correct = []float64{2, 3}

	res, err = Div(a, b)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())

	// tensor - tensor
	a = New(WithBacking([]float64{21, 10}))
	b = New(WithBacking([]float64{7, 2}))
	correct = []float64{3, 5}

	res, err = Div(a, b)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())
}

func TestAddScalarScalar(t *testing.T) {
	// scalar-scalar
	a := New(WithBacking([]float64{2}))
	b := New(WithBacking([]float64{3}))
	var correct interface{} = 5.0

	res, err := Add(a, b)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())

	// Test commutativity
	res, err = Add(b, a)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())

	// scalar-tensor
	a = New(WithBacking([]float64{3, 2}))
	b = New(WithBacking([]float64{2}))
	correct = []float64{5, 4}

	res, err = Add(a, b)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())

	// Test commutativity
	res, err = Add(b, a)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())

	// tensor - tensor
	a = New(WithBacking([]float64{3, 5}))
	b = New(WithBacking([]float64{7, 2}))
	correct = []float64{10, 7}

	res, err = Add(a, b)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())

	// Test commutativity
	res, err = Add(b, a)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())
}

func TestSubScalarScalar(t *testing.T) {
	// scalar-scalar
	a := New(WithBacking([]float64{6}))
	b := New(WithBacking([]float64{2}))
	var correct interface{} = 4.0

	res, err := Sub(a, b)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())

	// scalar-tensor
	a = New(WithBacking([]float64{6, 4}))
	b = New(WithBacking([]float64{2}))
	correct = []float64{4, 2}

	res, err = Sub(a, b)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())

	// tensor-scalar
	a = New(WithBacking([]float64{6}))
	b = New(WithBacking([]float64{3, 2}))
	correct = []float64{3, 4}

	res, err = Sub(a, b)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())

	// tensor - tensor
	a = New(WithBacking([]float64{21, 10}))
	b = New(WithBacking([]float64{7, 2}))
	correct = []float64{14, 8}

	res, err = Sub(a, b)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())
}

func TestModScalarScalar(t *testing.T) {
	// scalar-scalar
	a := New(WithBacking([]float64{5}))
	b := New(WithBacking([]float64{2}))
	var correct interface{} = 1.0

	res, err := Mod(a, b)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())

	// scalar-tensor
	a = New(WithBacking([]float64{5, 4}))
	b = New(WithBacking([]float64{2}))
	correct = []float64{1, 0}

	res, err = Mod(a, b)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())

	// tensor-scalar
	a = New(WithBacking([]float64{5}))
	b = New(WithBacking([]float64{3, 2}))
	correct = []float64{2, 1}

	res, err = Mod(a, b)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())

	// tensor - tensor
	a = New(WithBacking([]float64{22, 10}))
	b = New(WithBacking([]float64{7, 2}))
	correct = []float64{1, 0}

	res, err = Mod(a, b)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())
}

func TestPowScalarScalar(t *testing.T) {
	// scalar-scalar
	a := New(WithBacking([]float64{6}))
	b := New(WithBacking([]float64{2}))
	var correct interface{} = 36.0

	res, err := Pow(a, b)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())

	// scalar-tensor
	a = New(WithBacking([]float64{6, 4}))
	b = New(WithBacking([]float64{2}))
	correct = []float64{36, 16}

	res, err = Pow(a, b)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())

	// tensor-scalar
	a = New(WithBacking([]float64{6}))
	b = New(WithBacking([]float64{3, 2}))
	correct = []float64{216, 36}

	res, err = Pow(a, b)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())

	// tensor - tensor
	a = New(WithBacking([]float64{3, 10}))
	b = New(WithBacking([]float64{7, 2}))
	correct = []float64{2187, 100}

	res, err = Pow(a, b)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	assert.Equal(t, correct, res.Data())
}
