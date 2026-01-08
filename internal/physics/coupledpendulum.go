package physics

import (
	"math"

	"github.com/san-kum/dynsim/internal/dynamo"
)

// CoupledPendulums implements two pendulums connected by a spring.
// State: [theta1, omega1, theta2, omega2]
// Demonstrates energy transfer and coupled oscillations.
type CoupledPendulums struct {
	l float64 // Pendulum length
	g float64 // Gravity
	k float64 // Spring constant (coupling strength)
	m float64 // Mass of each bob
}

func NewCoupledPendulums() *CoupledPendulums {
	return &CoupledPendulums{
		l: 1.0,
		g: 9.81,
		k: 20.0, // Coupling spring constant
		m: 1.0,
	}
}

func (c *CoupledPendulums) StateDim() int   { return 4 }
func (c *CoupledPendulums) ControlDim() int { return 0 }

func (c *CoupledPendulums) Derive(state dynamo.State, _ dynamo.Control, _ float64) dynamo.State {
	theta1, omega1, theta2, omega2 := state[0], state[1], state[2], state[3]

	// Small angle approximation for simplicity
	// Coupling force proportional to (theta2 - theta1)
	coupling := c.k * (theta2 - theta1) / c.m

	// Pendulum 1: d²θ/dt² = -g/l * sin(θ) + coupling/l
	alpha1 := -c.g/c.l*math.Sin(theta1) + coupling/c.l

	// Pendulum 2: d²θ/dt² = -g/l * sin(θ) - coupling/l
	alpha2 := -c.g/c.l*math.Sin(theta2) - coupling/c.l

	return dynamo.State{omega1, alpha1, omega2, alpha2}
}

func (c *CoupledPendulums) DefaultState() dynamo.State {
	return dynamo.State{0.5, 0.0, 0.0, 0.0} // One pendulum displaced
}

// GetParams implements dynamo.Configurable
func (c *CoupledPendulums) GetParams() map[string]float64 {
	return map[string]float64{
		"l": c.l,
		"g": c.g,
		"k": c.k,
	}
}

// SetParam implements dynamo.Configurable
func (c *CoupledPendulums) SetParam(name string, value float64) {
	switch name {
	case "l":
		c.l = value
	case "g":
		c.g = value
	case "k":
		c.k = value
	}
}
