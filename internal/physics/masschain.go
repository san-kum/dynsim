package physics

import "github.com/san-kum/dynsim/internal/dynamo"

// MassChain implements a chain of masses connected by springs.
// Demonstrates wave propagation and standing waves.
// State: [x1, v1, x2, v2, ..., xN, vN] where x is displacement, v is velocity
type MassChain struct {
	n       int     // Number of masses
	k       float64 // Spring constant
	m       float64 // Mass of each particle
	damping float64 // Damping coefficient
}

func NewMassChain(n int) *MassChain {
	return &MassChain{
		n:       n,
		k:       100.0,
		m:       1.0,
		damping: 0.1,
	}
}

func (mc *MassChain) StateDim() int   { return mc.n * 2 }
func (mc *MassChain) ControlDim() int { return 0 }

func (mc *MassChain) Derive(state dynamo.State, _ dynamo.Control, _ float64) dynamo.State {
	deriv := make(dynamo.State, mc.n*2)

	for i := 0; i < mc.n; i++ {
		x := state[i*2]
		v := state[i*2+1]

		// Spring forces from neighbors
		var force float64

		// Left neighbor (or wall at x=0)
		if i > 0 {
			xLeft := state[(i-1)*2]
			force += mc.k * (xLeft - x)
		} else {
			force += mc.k * (0 - x) // Fixed wall at x=0
		}

		// Right neighbor (or wall)
		if i < mc.n-1 {
			xRight := state[(i+1)*2]
			force += mc.k * (xRight - x)
		} else {
			force += mc.k * (0 - x) // Fixed wall
		}

		// Damping
		force -= mc.damping * v

		// Acceleration
		a := force / mc.m

		deriv[i*2] = v
		deriv[i*2+1] = a
	}

	return deriv
}

func (mc *MassChain) DefaultState() dynamo.State {
	state := make(dynamo.State, mc.n*2)
	// Initial pulse - displace first few masses
	if mc.n > 0 {
		state[0] = 1.0 // First mass displaced
	}
	if mc.n > 2 {
		state[2] = 0.5 // Second mass
	}
	return state
}

// GetParams implements dynamo.Configurable
func (mc *MassChain) GetParams() map[string]float64 {
	return map[string]float64{
		"k":       mc.k,
		"damping": mc.damping,
	}
}

// SetParam implements dynamo.Configurable
func (mc *MassChain) SetParam(name string, value float64) {
	switch name {
	case "k":
		mc.k = value
	case "damping":
		mc.damping = value
	}
}
