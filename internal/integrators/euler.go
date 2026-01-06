package integrators

import "github.com/san-kum/dynsim/internal/sim"

type Euler struct{}

func NewEuler() *Euler {
	return &Euler{}
}

func (e *Euler) Step(dyn sim.Dynamics, x sim.State, u sim.Control, t float64, dt float64) sim.State {
	dx := dyn.Derivative(x, u, t)
	result := make(sim.State, len(x))
	for i := range x {
		result[i] = x[i] + dt*dx[i]
	}
	return result
}
