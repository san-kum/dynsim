package controllers

import (
	"testing"

	"github.com/san-kum/dynsim/internal/sim"
)

func TestNone(t *testing.T) {
	ctrl := NewNone(2)
	u := ctrl.Compute(sim.State{1.0, 2.0}, 0.0)

	if len(u) != 2 {
		t.Errorf("expected 2 controls, got %d", len(u))
	}
	for i, v := range u {
		if v != 0 {
			t.Errorf("control[%d] should be 0, got %f", i, v)
		}
	}
}

func TestPID(t *testing.T) {
	ctrl := NewPID(10.0, 0.1, 5.0, 0.0)
	u := ctrl.Compute(sim.State{1.0, 0.0}, 0.0)
	if len(u) != 1 {
		t.Fatalf("expected 1 control, got %d", len(u))
	}
	if u[0] >= 0 {
		t.Error("PID should output negative control for positive error")
	}
}

func TestLQR(t *testing.T) {
	k := [][]float64{{1.0, 2.0}}
	target := sim.State{0.0, 0.0}
	ctrl := NewLQR(k, target)

	u := ctrl.Compute(sim.State{0.0, 0.0}, 0.0)
	if u[0] != 0 {
		t.Errorf("expected zero control at target, got %f", u[0])
	}

	u = ctrl.Compute(sim.State{1.0, 0.0}, 0.0)
	if u[0] == 0 {
		t.Error("expected non-zero control away from target")
	}
}

func TestPendulumLQR(t *testing.T) {
	ctrl := NewPendulumLQR()
	u := ctrl.Compute(sim.State{0.1, 0.0}, 0.0)

	if len(u) != 1 {
		t.Fatalf("expected 1 control, got %d", len(u))
	}
	if u[0] == 0 {
		t.Error("pendulum LQR should output non-zero control for non-zero angle")
	}
}

func TestCartPoleLQR(t *testing.T) {
	ctrl := NewCartPoleLQR()
	u := ctrl.Compute(sim.State{0.0, 0.0, 0.1, 0.0}, 0.0)

	if len(u) != 1 {
		t.Fatalf("expected 1 control, got %d", len(u))
	}
}
