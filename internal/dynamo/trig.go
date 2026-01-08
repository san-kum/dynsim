package dynamo

import "math"

// TrigTable provides precomputed sin/cos values for fast lookup.
// Uses linear interpolation for values between table entries.
type TrigTable struct {
	sin []float64
	cos []float64
	n   int
}

// Global default trig table (4096 entries = ~0.0015 rad resolution)
var DefaultTrigTable = NewTrigTable(4096)

// NewTrigTable creates a precomputed trig lookup table
func NewTrigTable(n int) *TrigTable {
	t := &TrigTable{
		sin: make([]float64, n),
		cos: make([]float64, n),
		n:   n,
	}

	for i := 0; i < n; i++ {
		angle := float64(i) * 2 * math.Pi / float64(n)
		t.sin[i] = math.Sin(angle)
		t.cos[i] = math.Cos(angle)
	}

	return t
}

// Sin returns approximate sin using table lookup with interpolation
func (t *TrigTable) Sin(x float64) float64 {
	// Normalize to [0, 2π)
	x = math.Mod(x, 2*math.Pi)
	if x < 0 {
		x += 2 * math.Pi
	}

	// Map to table index
	idx := x * float64(t.n) / (2 * math.Pi)
	i := int(idx)
	frac := idx - float64(i)

	// Linear interpolation
	i0 := i % t.n
	i1 := (i + 1) % t.n

	return t.sin[i0]*(1-frac) + t.sin[i1]*frac
}

// Cos returns approximate cos using table lookup with interpolation
func (t *TrigTable) Cos(x float64) float64 {
	x = math.Mod(x, 2*math.Pi)
	if x < 0 {
		x += 2 * math.Pi
	}

	idx := x * float64(t.n) / (2 * math.Pi)
	i := int(idx)
	frac := idx - float64(i)

	i0 := i % t.n
	i1 := (i + 1) % t.n

	return t.cos[i0]*(1-frac) + t.cos[i1]*frac
}

// SinCos returns both sin and cos efficiently
func (t *TrigTable) SinCos(x float64) (sin, cos float64) {
	x = math.Mod(x, 2*math.Pi)
	if x < 0 {
		x += 2 * math.Pi
	}

	idx := x * float64(t.n) / (2 * math.Pi)
	i := int(idx)
	frac := idx - float64(i)

	i0 := i % t.n
	i1 := (i + 1) % t.n

	sin = t.sin[i0]*(1-frac) + t.sin[i1]*frac
	cos = t.cos[i0]*(1-frac) + t.cos[i1]*frac
	return
}

// FastSin uses the default table for quick sin lookup
func FastSin(x float64) float64 {
	return DefaultTrigTable.Sin(x)
}

// FastCos uses the default table for quick cos lookup
func FastCos(x float64) float64 {
	return DefaultTrigTable.Cos(x)
}

// FastSinCos uses the default table
func FastSinCos(x float64) (float64, float64) {
	return DefaultTrigTable.SinCos(x)
}
