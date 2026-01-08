package physics

import (
	"fmt"
	"math"

	"github.com/san-kum/dynsim/internal/dynamo"
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

func (p *Pendulum) Derive(x dynamo.State, u dynamo.Control, t float64) dynamo.State {
	theta := x[0]
	omega := x[1]

	torque := 0.0
	if len(x) > 0 {
		torque = u[0]
	}
	alpha := (-p.Damping*omega - p.Mass*p.Gravity*p.Length*math.Sin(theta) + torque) / (p.Mass * p.Length * p.Length)

	return dynamo.State{omega, alpha}
}

func (p *Pendulum) Energy(x dynamo.State) float64 {
	// KE = 0.5 * m * (L*omega)^2
	// PE = m * g * L * (1 - cos(theta))
	v := p.Length * x[1]
	ke := 0.5 * p.Mass * v * v
	pe := p.Mass * p.Gravity * p.Length * (1.0 - math.Cos(x[0]))
	return ke + pe
}

func (p *Pendulum) GetParams() map[string]float64 {
	return map[string]float64{
		"mass":    p.Mass,
		"length":  p.Length,
		"damping": p.Damping,
		"gravity": p.Gravity,
	}
}

func (p *Pendulum) SetParam(name string, value float64) error {
	switch name {
	case "mass":
		p.Mass = value
	case "length":
		p.Length = value
	case "damping":
		p.Damping = value
	case "gravity":
		p.Gravity = value
	default:
		return fmt.Errorf("unknown param: %s", name)
	}
	return nil
}
