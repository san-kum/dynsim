package models

import (
	"math"
	"testing"

	"github.com/san-kum/dynsim/internal/sim"
)

func TestDroneStateDim(t *testing.T) {
	d := NewDrone()
	if d.StateDim() != 6 {
		t.Errorf("expected 6 states, got %d", d.StateDim())
	}
	if d.ControlDim() != 2 {
		t.Errorf("expected 2 controls, got %d", d.ControlDim())
	}
}

func TestDroneHover(t *testing.T) {
	d := NewDrone()
	hoverThrust := d.HoverThrust()

	x := sim.State{0, 5, 0, 0, 0, 0}
	u := sim.Control{hoverThrust, hoverThrust}

	dx := d.Derivative(x, u, 0.0)

	if math.Abs(dx[4]) > 0.01 {
		t.Errorf("vertical acceleration should be ~0, got %f", dx[4])
	}

	if math.Abs(dx[3]) > 0.01 {
		t.Errorf("horizontal acceleration should be ~0, got %f", dx[3])
	}

	if math.Abs(dx[5]) > 0.01 {
		t.Errorf("angular acceleration should be ~0, got %f", dx[5])
	}
}

func TestDroneFreefall(t *testing.T) {
	d := NewDrone()

	x := sim.State{0, 5, 0, 0, 0, 0}
	u := sim.Control{0, 0}

	dx := d.Derivative(x, u, 0.0)

	expectedAy := -d.Gravity
	if math.Abs(dx[4]-expectedAy) > 0.1 {
		t.Errorf("expected ay=%f, got %f", expectedAy, dx[4])
	}
}

func TestDroneTorque(t *testing.T) {
	d := NewDrone()

	x := sim.State{0, 5, 0, 0, 0, 0}
	u := sim.Control{0, 5}

	dx := d.Derivative(x, u, 0.0)

	if dx[5] <= 0 {
		t.Errorf("angular acceleration should be positive, got %f", dx[5])
	}
}

func TestDroneEnergy(t *testing.T) {
	d := NewDrone()
	x := sim.State{0, 10, 0, 1, 2, 0.5}

	e := d.Energy(x)
	if e <= 0 {
		t.Error("energy should be positive")
	}
}
