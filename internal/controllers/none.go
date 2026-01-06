package controllers

import "github.com/san-kum/dynsim/internal/sim"

type None struct {
	dim int
}

func NewNone(dim int) *None {
	return &None{
		dim: dim,
	}
}

func (n *None) Compute(x sim.State, t float64) sim.Control {
	return make(sim.Control, n.dim)
}
