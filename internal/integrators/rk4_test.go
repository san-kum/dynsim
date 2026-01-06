package integrators

import (
	"math"
	"testing"

	"github.com/san-kum/dynsim/internal/sim"
)

type simpleDynamics struct{}

func (s *simpleDynamics) Derivative(x sim.State, u sim.Control, t float64) sim.State {
	return sim.State{x[1], -x[0]}
}

func (s *simpleDynamics) StateDim() int   { return 2 }
func (s *simpleDynamics) ControlDim() int { return 0 }

func TestRK4Accuracy(t *testing.T) {
	dyn := &simpleDynamics{}
	integ := NewRK4()

	x0 := sim.State{1.0, 0.0}
	u := sim.Control{}
	dt := 0.01
	steps := 100

	x := x0
	for i := 0; i < steps; i++ {
		x = integ.Step(dyn, x, u, float64(i)*dt, dt)
	}

	expectedX := math.Cos(float64(steps) * dt)
	expectedV := -math.Sin(float64(steps) * dt)

	if math.Abs(x[0]-expectedX) > 1e-4 {
		t.Errorf("position error too large: got %.6f, expected %.6f", x[0], expectedX)
	}

	if math.Abs(x[1]-expectedV) > 1e-4 {
		t.Errorf("velocity error too large: got %.6f, expected %.6f", x[1], expectedV)
	}
}
