package integrators

import "github.com/san-kum/dynsim/internal/sim"

type RK4 struct{}

func NewRK4() *RK4 {
	return &RK4{}
}

func (r *RK4) Step(dyn sim.Dynamics, x sim.State, u sim.Control, t float64, dt float64) sim.State {

	k1 := dyn.Derivative(x, u, t)
	x2 := addScaled(x, k1, dt/2)
	k2 := dyn.Derivative(x2, u, t+dt/2)

	x3 := addScaled(x, k2, dt/2)
	k3 := dyn.Derivative(x3, u, t+dt/2)

	x4 := addScaled(x, k3, dt)
	k4 := dyn.Derivative(x4, u, t+dt)

	result := make(sim.State, len(x))
	for i := range x {
		result[i] = x[i] + (dt/6)*(k1[i]+2*k2[i]+2*k3[i]+k4[i])
	}

	return result
}

func addScaled(x sim.State, dx sim.State, scale float64) sim.State {
	result := make(sim.State, len(x))
	for i := range x {
		result[i] = x[i] + scale*dx[i]
	}
	return result
}
