package models

import (
	"math"
	"testing"

	"github.com/san-kum/dynsim/internal/sim"
)

func TestDoublePendulumEquilibrium(t *testing.T) {
	dp := NewDoublePendulum()

	// At rest hanging straight down
	x := sim.State{0, 0, 0, 0}
	u := sim.Control{0}

	dx := dp.Derivative(x, u, 0)

	// Velocities should be zero
	if math.Abs(dx[0]) > 1e-10 {
		t.Errorf("expected zero omega1, got %f", dx[0])
	}
	if math.Abs(dx[1]) > 1e-10 {
		t.Errorf("expected zero omega2, got %f", dx[1])
	}
	// Accelerations should be zero at equilibrium
	if math.Abs(dx[2]) > 1e-10 {
		t.Errorf("expected zero alpha1, got %f", dx[2])
	}
	if math.Abs(dx[3]) > 1e-10 {
		t.Errorf("expected zero alpha2, got %f", dx[3])
	}
}

func TestDoublePendulumDimensions(t *testing.T) {
	dp := NewDoublePendulum()

	if dp.StateDim() != 4 {
		t.Errorf("expected state dim 4, got %d", dp.StateDim())
	}
	if dp.ControlDim() != 1 {
		t.Errorf("expected control dim 1, got %d", dp.ControlDim())
	}
}

func TestDoublePendulumSymmetry(t *testing.T) {
	dp := NewDoublePendulum()

	// Symmetric initial condition should give symmetric accelerations
	x1 := sim.State{0.1, 0.1, 0, 0}
	x2 := sim.State{-0.1, -0.1, 0, 0}
	u := sim.Control{0}

	dx1 := dp.Derivative(x1, u, 0)
	dx2 := dp.Derivative(x2, u, 0)

	// Accelerations should be opposite
	if math.Abs(dx1[2]+dx2[2]) > 1e-6 {
		t.Errorf("expected symmetric alpha1: %f vs %f", dx1[2], dx2[2])
	}
	if math.Abs(dx1[3]+dx2[3]) > 1e-6 {
		t.Errorf("expected symmetric alpha2: %f vs %f", dx1[3], dx2[3])
	}
}
