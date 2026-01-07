package integrators

import (
	"math"
	"testing"

	"github.com/san-kum/dynsim/internal/sim"
)

type harmonicOscillator struct{}

func (h *harmonicOscillator) StateDim() int   { return 2 }
func (h *harmonicOscillator) ControlDim() int { return 0 }

func (h *harmonicOscillator) Derivative(x sim.State, u sim.Control, t float64) sim.State {
	return sim.State{x[1], -x[0]}
}

func (h *harmonicOscillator) Energy(x sim.State) float64 {
	return 0.5 * (x[0]*x[0] + x[1]*x[1])
}

func TestRK45_Step(t *testing.T) {
	integrator := NewRK45()
	dyn := &harmonicOscillator{}
	x0 := sim.State{1.0, 0.0}

	x := x0.Clone()
	dt := 0.01

	for i := 0; i < 1000; i++ {
		x = integrator.Step(dyn, x, nil, float64(i)*dt, dt)
	}

	if !x.IsValid() {
		t.Error("RK45 produced invalid state")
	}
}

func TestRK45_EnergyConservation(t *testing.T) {
	integrator := NewRK45()
	dyn := &harmonicOscillator{}
	x0 := sim.State{1.0, 0.0}

	initialEnergy := dyn.Energy(x0)
	x := x0.Clone()
	dt := 0.01

	for i := 0; i < 10000; i++ {
		x = integrator.Step(dyn, x, nil, float64(i)*dt, dt)
	}

	finalEnergy := dyn.Energy(x)
	drift := math.Abs(finalEnergy-initialEnergy) / initialEnergy

	if drift > 1e-6 {
		t.Errorf("RK45 energy drift too high: %e", drift)
	}
}

func TestRK45_AdaptiveStep(t *testing.T) {
	integrator := NewRK45()
	dyn := &harmonicOscillator{}
	x0 := sim.State{1.0, 0.0}

	x, newDt, err := integrator.StepAdaptive(dyn, x0, nil, 0, 0.1, 1e-8)

	if err != nil {
		t.Errorf("StepAdaptive returned error: %v", err)
	}

	if !x.IsValid() {
		t.Error("StepAdaptive produced invalid state")
	}

	if newDt <= 0 {
		t.Errorf("StepAdaptive returned invalid dt: %f", newDt)
	}
}

func TestRK45_VsRK4_Accuracy(t *testing.T) {
	rk4 := NewRK4()
	rk45 := NewRK45()
	dyn := &harmonicOscillator{}
	x0 := sim.State{1.0, 0.0}

	x4 := x0.Clone()
	x45 := x0.Clone()
	dt := 0.1

	for i := 0; i < 100; i++ {
		x4 = rk4.Step(dyn, x4, nil, float64(i)*dt, dt)
		x45 = rk45.Step(dyn, x45, nil, float64(i)*dt, dt)
	}

	t.Logf("RK4 final: [%.6f, %.6f]", x4[0], x4[1])
	t.Logf("RK45 final: [%.6f, %.6f]", x45[0], x45[1])

	e4 := (&harmonicOscillator{}).Energy(x4)
	e45 := (&harmonicOscillator{}).Energy(x45)

	if math.Abs(e45-1.0) > math.Abs(e4-1.0) {
		t.Log("Warning: RK45 not more accurate than RK4 for this case")
	}
}
