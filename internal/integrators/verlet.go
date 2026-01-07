package integrators

import "github.com/san-kum/dynsim/internal/sim"

type Verlet struct {
	prevAcc sim.State
	scratch sim.State
}

func NewVerlet() *Verlet {
	return &Verlet{}
}

func (v *Verlet) ensureScratch(n int) {
	if len(v.scratch) != n {
		v.scratch = make(sim.State, n)
		v.prevAcc = nil
	}
}

func (v *Verlet) Step(dyn sim.Dynamics, x sim.State, u sim.Control, t, dt float64) sim.State {
	n := len(x)
	half := n / 2
	v.ensureScratch(n)

	result := make(sim.State, n)
	dx := dyn.Derivative(x, u, t)
	dt2 := dt * dt

	for i := 0; i < half; i++ {
		result[i] = x[i] + x[half+i]*dt + 0.5*dx[half+i]*dt2
	}

	for i := 0; i < half; i++ {
		v.scratch[i] = result[i]
		v.scratch[half+i] = x[half+i]
	}

	dxNew := dyn.Derivative(v.scratch, u, t+dt)

	halfDt := 0.5 * dt
	for i := 0; i < half; i++ {
		result[half+i] = x[half+i] + (dx[half+i]+dxNew[half+i])*halfDt
	}

	return result
}

type Leapfrog struct {
	scratch sim.State
}

func NewLeapfrog() *Leapfrog {
	return &Leapfrog{}
}

func (l *Leapfrog) Step(dyn sim.Dynamics, x sim.State, u sim.Control, t, dt float64) sim.State {
	n := len(x)
	half := n / 2

	if len(l.scratch) != n {
		l.scratch = make(sim.State, n)
	}

	result := make(sim.State, n)
	dx := dyn.Derivative(x, u, t)
	halfDt := dt * 0.5

	for i := 0; i < half; i++ {
		l.scratch[half+i] = x[half+i] + dx[half+i]*halfDt
	}

	for i := 0; i < half; i++ {
		result[i] = x[i] + l.scratch[half+i]*dt
		l.scratch[i] = result[i]
	}

	dxNew := dyn.Derivative(l.scratch, u, t+dt)

	for i := 0; i < half; i++ {
		result[half+i] = l.scratch[half+i] + dxNew[half+i]*halfDt
	}

	return result
}
