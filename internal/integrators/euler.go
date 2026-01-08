package integrators

import "github.com/san-kum/dynsim/internal/dynamo"

type Euler struct{}

func NewEuler() *Euler {
	return &Euler{}
}

func (e *Euler) Step(dyn dynamo.System, x dynamo.State, u dynamo.Control, t float64, dt float64) dynamo.State {
	dx := dyn.Derive(x, u, t)
	result := make(dynamo.State, len(x))
	for i := range x {
		result[i] = x[i] + dt*dx[i]
	}
	return result
}
