package metrics

import (
	"math"
	"testing"

	"github.com/san-kum/dynsim/internal/sim"
)

func TestEnergyConservation(t *testing.T) {
	m := NewEnergy(1.0, 1.0, 9.81)

	theta := math.Pi / 4
	omega := 0.0

	x := sim.State{theta, omega}
	u := sim.Control{}

	m.Observe(x, u, 0)
	e1 := m.Value()

	m.Reset()

	ke := 0.5 * omega * omega
	pe := 9.81 * (1 - math.Cos(theta))
	expected := ke + pe

	m.Observe(x, u, 0)
	e2 := m.Value()

	if math.Abs(e1-expected) > 1e-6 {
		t.Errorf("expected energy %f, got %f", expected, e1)
	}

	if math.Abs(e2-expected) > 1e-6 {
		t.Errorf("expected energy %f after reset, got %f", expected, e2)
	}
}

func TestEnergyReset(t *testing.T) {
	m := NewEnergy(1.0, 1.0, 9.81)

	x := sim.State{1.0, 1.0}
	u := sim.Control{}

	m.Observe(x, u, 0)
	if m.Value() == 0 {
		t.Error("expected non-zero energy")
	}

	m.Reset()
	if m.Value() != 0 {
		t.Error("expected zero energy after reset")
	}
}
