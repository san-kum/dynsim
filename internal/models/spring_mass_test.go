package models

import (
	"math"
	"testing"

	"github.com/san-kum/dynsim/internal/sim"
)

func TestSpringMassDerivative_Equilibrium(t *testing.T) {
	sm := NewSpringMass()
	x := sim.State{0.0, 0.0}
	u := sim.Control{0.0}

	dx := sm.Derivative(x, u, 0.0)

	if dx[0] != 0 {
		t.Errorf("velocity at equilibrium should be 0, got %f", dx[0])
	}
	if dx[1] != 0 {
		t.Errorf("acceleration at equilibrium should be 0, got %f", dx[1])
	}
}

func TestSpringMassDerivative_Displaced(t *testing.T) {
	sm := NewSpringMass()
	x := sim.State{1.0, 0.0}
	u := sim.Control{0.0}

	dx := sm.Derivative(x, u, 0.0)

	if dx[0] != 0 {
		t.Errorf("velocity should be 0, got %f", dx[0])
	}

	expectedAcc := -DefaultStiffness * 1.0 / DefaultMass
	if math.Abs(dx[1]-expectedAcc) > 0.001 {
		t.Errorf("expected acceleration %f, got %f", expectedAcc, dx[1])
	}
}

func TestSpringMassEnergy(t *testing.T) {
	sm := NewSpringMass()

	x := sim.State{1.0, 0.0}
	e1 := sm.Energy(x)

	x = sim.State{0.0, 3.16}
	e2 := sm.Energy(x)

	if math.Abs(e1-e2) > 1.0 {
		t.Errorf("energy should be approximately conserved: PE=%f, KE=%f", e1, e2)
	}
}

func TestSpringMassChain(t *testing.T) {
	chain := NewSpringMassChain(3)

	if chain.StateDim() != 6 {
		t.Errorf("expected 6 states, got %d", chain.StateDim())
	}

	x := sim.State{0.0, 0.0, 0.0, 0.0, 0.0, 0.0}
	dx := chain.Derivative(x, nil, 0.0)

	for i := 0; i < 6; i++ {
		if dx[i] != 0 {
			t.Errorf("derivative[%d] at equilibrium should be 0, got %f", i, dx[i])
		}
	}
}
