package integrators

import "github.com/san-kum/dynsim/internal/sim"

type Verlet struct{}

func NewVerlet() *Verlet {
	return &Verlet{}
}

func (v *Verlet) Step(dyn sim.Dynamics, x sim.State, u sim.Control, t float64, dt float64) sim.State {
	n := len(x) / 2
	result := make(sim.State, len(x))

	dx := dyn.Derivative(x, u, t)

	for i := 0; i < n; i++ {
		result[i] = x[i] + x[n+i]*dt + 0.5*dx[n+i]*dt*dt
	}

	xmid := make(sim.State, len(x))
	copy(xmid, result)
	for i := 0; i < n; i++ {
		xmid[n+i] = x[n+i]
	}

	dxmid := dyn.Derivative(xmid, u, t+dt)

	for i := 0; i < n; i++ {
		result[n+i] = x[n+i] + 0.5*(dx[n+i]+dxmid[n+i])*dt
	}

	return result
}
