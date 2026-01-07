package integrators

import "github.com/san-kum/dynsim/internal/sim"

type RK4 struct {
	k1, k2, k3, k4 sim.State
	scratch        sim.State
}

func NewRK4() *RK4 {
	return &RK4{}
}

func (r *RK4) ensureScratch(n int) {
	if len(r.k1) != n {
		r.k1 = make(sim.State, n)
		r.k2 = make(sim.State, n)
		r.k3 = make(sim.State, n)
		r.k4 = make(sim.State, n)
		r.scratch = make(sim.State, n)
	}
}

func (r *RK4) Step(dyn sim.Dynamics, x sim.State, u sim.Control, t, dt float64) sim.State {
	n := len(x)
	r.ensureScratch(n)

	k1 := dyn.Derivative(x, u, t)
	copy(r.k1, k1)

	for i := 0; i < n; i++ {
		r.scratch[i] = x[i] + dt*0.5*r.k1[i]
	}
	k2 := dyn.Derivative(r.scratch, u, t+dt*0.5)
	copy(r.k2, k2)

	for i := 0; i < n; i++ {
		r.scratch[i] = x[i] + dt*0.5*r.k2[i]
	}
	k3 := dyn.Derivative(r.scratch, u, t+dt*0.5)
	copy(r.k3, k3)

	for i := 0; i < n; i++ {
		r.scratch[i] = x[i] + dt*r.k3[i]
	}
	k4 := dyn.Derivative(r.scratch, u, t+dt)
	copy(r.k4, k4)

	result := make(sim.State, n)
	dt6 := dt / 6.0
	for i := 0; i < n; i++ {
		result[i] = x[i] + dt6*(r.k1[i]+2*r.k2[i]+2*r.k3[i]+r.k4[i])
	}

	return result
}
