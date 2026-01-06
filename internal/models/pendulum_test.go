package models

import (
	"math"
	"testing"

	"github.com/san-kum/dynsim/internal/sim"
)

func TestPendulumEquilibrium(t *testing.T) {
	p := NewPendulum()
	p.Damping = 0

	x := sim.State{0, 0}
	u := sim.Control{0}

	dx := p.Derivative(x, u, 0)

	if math.Abs(dx[0]) > 1e-10 {
		t.Errorf("expected zero velocity at equilibrium, got %f", dx[0])
	}

	if math.Abs(dx[1]) > 1e-10 {
		t.Errorf("expected zero acceleration at equilibrium, got %f", dx[1])
	}
}

func TestPendulumDimensions(t *testing.T) {
	p := NewPendulum()

	if p.StateDim() != 2 {
		t.Errorf("expected state dim 2, got %d", p.StateDim())
	}

	if p.ControlDim() != 1 {
		t.Errorf("expected control dim 1, got %d", p.ControlDim())
	}
}

func TestPendulumGravity(t *testing.T) {
	p := NewPendulum()
	p.Damping = 0

	x := sim.State{math.Pi / 2, 0}
	u := sim.Control{0}

	dx := p.Derivative(x, u, 0)

	expectedAccel := -p.Gravity / p.Length

	if math.Abs(dx[1]-expectedAccel) > 1e-6 {
		t.Errorf("expected acceleration %f, got %f", expectedAccel, dx[1])
	}
}
