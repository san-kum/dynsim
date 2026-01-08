package physics

import (
	"math"

	"github.com/san-kum/dynsim/internal/dynamo"
)

// DoubleWell models a particle in a bistable potential well.
type DoubleWell struct {
	A, B, Mass, Damping float64
}

func NewDoubleWell() *DoubleWell {
	return &DoubleWell{1.0, 1.0, 1.0, 0.1}
}

func (d *DoubleWell) StateDim() int   { return 2 }
func (d *DoubleWell) ControlDim() int { return 1 }

func (d *DoubleWell) Derive(s dynamo.State, u dynamo.Control, _ float64) dynamo.State {
	if len(s) < 2 {
		return make(dynamo.State, 2)
	}
	x, v := s[0], s[1]
	ef := 0.0
	if len(u) > 0 {
		ef = u[0]
	}
	return dynamo.State{v, (-4*d.A*x*(x*x-d.B) - d.Damping*v + ef) / d.Mass}
}

func (d *DoubleWell) DefaultState() dynamo.State { return dynamo.State{math.Sqrt(d.B) + 0.1, 0} }

func (d *DoubleWell) Energy(s dynamo.State) float64 {
	if len(s) < 2 {
		return 0
	}
	x, v := s[0], s[1]
	return 0.5*d.Mass*v*v + d.A*math.Pow(x*x-d.B, 2)
}

func (d *DoubleWell) GetParams() map[string]float64 {
	return map[string]float64{"A": d.A, "B": d.B, "mass": d.Mass, "damping": d.Damping}
}

func (d *DoubleWell) SetParam(n string, v float64) error {
	switch n {
	case "A":
		d.A = v
	case "B":
		d.B = v
	case "mass":
		d.Mass = v
	case "damping":
		d.Damping = v
	}
	return nil
}
