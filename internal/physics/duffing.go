package physics

import (
	"math"

	"github.com/san-kum/dynsim/internal/dynamo"
)

// Duffing implements a nonlinear forced oscillator.
type Duffing struct {
	Alpha, Beta, Delta, Gamma, Omega float64
}

func NewDuffing() *Duffing {
	return &Duffing{-1.0, 1.0, 0.3, 0.5, 1.2}
}

func (d *Duffing) StateDim() int   { return 3 }
func (d *Duffing) ControlDim() int { return 0 }

func (d *Duffing) Derive(s dynamo.State, _ dynamo.Control, _ float64) dynamo.State {
	if len(s) < 3 {
		return make(dynamo.State, 3)
	}
	x, v, phi := s[0], s[1], s[2]
	return dynamo.State{v, -d.Delta*v - d.Alpha*x - d.Beta*x*x*x + d.Gamma*math.Cos(phi), d.Omega}
}

func (d *Duffing) DefaultState() dynamo.State { return dynamo.State{1.0, 0.0, 0.0} }

func (d *Duffing) Energy(s dynamo.State) float64 {
	if len(s) < 2 {
		return 0
	}
	x, v := s[0], s[1]
	return 0.5*v*v + 0.5*d.Alpha*x*x + 0.25*d.Beta*x*x*x*x
}

func (d *Duffing) LyapunovExponent() float64 {
	if d.Gamma > 0.3 && d.Gamma < 0.65 && d.Delta < 0.5 {
		return 0.15
	}
	return 0.0
}

func (d *Duffing) GetParams() map[string]float64 {
	return map[string]float64{"alpha": d.Alpha, "beta": d.Beta, "delta": d.Delta, "gamma": d.Gamma, "omega": d.Omega}
}

func (d *Duffing) SetParam(n string, v float64) error {
	switch n {
	case "alpha":
		d.Alpha = v
	case "beta":
		d.Beta = v
	case "delta":
		d.Delta = v
	case "gamma":
		d.Gamma = v
	case "omega":
		d.Omega = v
	}
	return nil
}
