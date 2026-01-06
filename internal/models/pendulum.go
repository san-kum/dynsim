package models

import (
	"math"

	"github.com/san-kum/dynsim/internal/sim"
)

type Pendulum struct {
	Mass    float64
	Length  float64
	Damping float64
	Gravity float64
}

func NewPendulum() *Pendulum {
	return &Pendulum{
		Mass:    1.0,
		Length:  1.0,
		Damping: 0.1,
		Gravity: 9.81,
	}
}

func (p *Pendulum) StateDim() int {
	return 2
}

func (p *Pendulum) ControlDim() int {
	return 1
}

func (p *Pendulum) Derivative(x sim.State, u sim.Control, t float64) sim.State {
	theta := x[0]
	omega := x[1]

	torque := 0.0
	if len(x) > 0 {
		torque = u[0]
	}
	alpha := (-p.Damping*omega - p.Mass*p.Gravity*p.Length*math.Sin(theta) + torque) / (p.Mass * p.Length * p.Length)

	return sim.State{omega, alpha}
}
