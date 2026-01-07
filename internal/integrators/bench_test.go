package integrators

import (
	"testing"

	"github.com/san-kum/dynsim/internal/sim"
)

type benchDynamics struct{}

func (b *benchDynamics) StateDim() int   { return 2 }
func (b *benchDynamics) ControlDim() int { return 0 }
func (b *benchDynamics) Derivative(x sim.State, u sim.Control, t float64) sim.State {
	return sim.State{x[1], -x[0]}
}

func BenchmarkEuler(b *testing.B) {
	integrator := NewEuler()
	dyn := &benchDynamics{}
	x := sim.State{1.0, 0.0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x = integrator.Step(dyn, x, nil, 0, 0.01)
	}
}

func BenchmarkRK4(b *testing.B) {
	integrator := NewRK4()
	dyn := &benchDynamics{}
	x := sim.State{1.0, 0.0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x = integrator.Step(dyn, x, nil, 0, 0.01)
	}
}

func BenchmarkRK45(b *testing.B) {
	integrator := NewRK45()
	dyn := &benchDynamics{}
	x := sim.State{1.0, 0.0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x = integrator.Step(dyn, x, nil, 0, 0.01)
	}
}

func BenchmarkVerlet(b *testing.B) {
	integrator := NewVerlet()
	dyn := &benchDynamics{}
	x := sim.State{1.0, 0.0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x = integrator.Step(dyn, x, nil, 0, 0.01)
	}
}

func BenchmarkLeapfrog(b *testing.B) {
	integrator := NewLeapfrog()
	dyn := &benchDynamics{}
	x := sim.State{1.0, 0.0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x = integrator.Step(dyn, x, nil, 0, 0.01)
	}
}

type benchNBody struct{}

func (b *benchNBody) StateDim() int   { return 20 }
func (b *benchNBody) ControlDim() int { return 0 }
func (b *benchNBody) Derivative(x sim.State, u sim.Control, t float64) sim.State {
	dx := make(sim.State, 20)
	for i := 0; i < 5; i++ {
		dx[i*4] = x[i*4+2]
		dx[i*4+1] = x[i*4+3]
		dx[i*4+2] = -x[i*4] * 0.1
		dx[i*4+3] = -x[i*4+1] * 0.1
	}
	return dx
}

func BenchmarkRK4_NBody5(b *testing.B) {
	integrator := NewRK4()
	dyn := &benchNBody{}
	x := make(sim.State, 20)
	for i := range x {
		x[i] = float64(i) * 0.1
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x = integrator.Step(dyn, x, nil, 0, 0.001)
	}
}

func BenchmarkLeapfrog_NBody5(b *testing.B) {
	integrator := NewLeapfrog()
	dyn := &benchNBody{}
	x := make(sim.State, 20)
	for i := range x {
		x[i] = float64(i) * 0.1
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x = integrator.Step(dyn, x, nil, 0, 0.001)
	}
}
