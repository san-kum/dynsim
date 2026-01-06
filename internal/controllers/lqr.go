package controllers

import "github.com/san-kum/dynsim/internal/sim"

type LQR struct {
	K      [][]float64
	Target sim.State
}

func NewLQR(k [][]float64, target sim.State) *LQR {
	return &LQR{K: k, Target: target}
}

func (l *LQR) Compute(x sim.State, t float64) sim.Control {
	u := make(sim.Control, len(l.K))

	for i := range u {
		for j := range x {
			target := 0.0
			if j < len(l.Target) {
				target = l.Target[j]
			}
			u[i] -= l.K[i][j] * (x[j] - target)
		}
	}

	return u
}

func NewPendulumLQR() *LQR {
	k := [][]float64{
		{31.62, 10.0},
	}
	return NewLQR(k, sim.State{0, 0})
}
