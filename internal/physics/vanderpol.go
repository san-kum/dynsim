package physics

import "github.com/san-kum/dynsim/internal/dynamo"

// VanDerPol implements the Van der Pol oscillator.
// State: [x, y] where y = dx/dt
// Equations:
//
//	dx/dt = y
//	dy/dt = μ(1 - x²)y - x
type VanDerPol struct {
	mu float64 // Nonlinearity parameter
}

func NewVanDerPol() *VanDerPol {
	return &VanDerPol{
		mu: 1.0, // Classic value for limit cycle
	}
}

func (v *VanDerPol) StateDim() int   { return 2 }
func (v *VanDerPol) ControlDim() int { return 0 }

func (v *VanDerPol) Derive(state dynamo.State, _ dynamo.Control, _ float64) dynamo.State {
	x, y := state[0], state[1]

	dx := y
	dy := v.mu*(1-x*x)*y - x

	return dynamo.State{dx, dy}
}

func (v *VanDerPol) DefaultState() dynamo.State {
	return dynamo.State{2.0, 0.0}
}

// GetParams implements dynamo.Configurable
func (v *VanDerPol) GetParams() map[string]float64 {
	return map[string]float64{
		"mu": v.mu,
	}
}

// SetParam implements dynamo.Configurable
func (v *VanDerPol) SetParam(name string, value float64) {
	if name == "mu" {
		v.mu = value
	}
}
