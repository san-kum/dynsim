package physics

import "github.com/san-kum/dynsim/internal/dynamo"

type Lorenz struct{ sigma, rho, beta float64 }

func NewLorenz() *Lorenz          { return &Lorenz{10.0, 28.0, 8.0 / 3.0} }
func (l *Lorenz) StateDim() int   { return 3 }
func (l *Lorenz) ControlDim() int { return 0 }

// Derive calculates the Lorenz attractor derivatives.
func (l *Lorenz) Derive(s dynamo.State, _ dynamo.Control, _ float64) dynamo.State {
	return dynamo.State{l.sigma * (s[1] - s[0]), s[0]*(l.rho-s[2]) - s[1], s[0]*s[1] - l.beta*s[2]}
}
func (l *Lorenz) DefaultState() dynamo.State { return dynamo.State{1.0, 1.0, 1.0} }
func (l *Lorenz) GetParams() map[string]float64 {
	return map[string]float64{"sigma": l.sigma, "rho": l.rho, "beta": l.beta}
}
func (l *Lorenz) SetParam(n string, v float64) {
	switch n {
	case "sigma":
		l.sigma = v
	case "rho":
		l.rho = v
	case "beta":
		l.beta = v
	}
}
