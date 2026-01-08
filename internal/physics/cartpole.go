package physics

import (
	"math"

	"github.com/san-kum/dynsim/internal/dynamo"
)

type CartPole struct {
	CartMass   float64
	PoleMass   float64
	PoleLength float64
	Gravity    float64
}

func NewCartPole() *CartPole {
	return &CartPole{
		CartMass:   1.0,
		PoleMass:   0.1,
		PoleLength: 1.0,
		Gravity:    9.81,
	}
}

func (c *CartPole) StateDim() int {
	return 4
}

func (c *CartPole) ControlDim() int {
	return 1
}

func (c *CartPole) Derive(x dynamo.State, u dynamo.Control, t float64) dynamo.State {
	pos := x[0]
	vel := x[1]
	theta := x[2]
	omega := x[3]

	force := 0.0
	if len(u) > 0 {
		force = u[0]
	}

	_ = pos

	mc := c.CartMass
	mp := c.PoleMass
	l := c.PoleLength
	g := c.Gravity

	sint := math.Sin(theta)
	cost := math.Cos(theta)

	temp := (force + mp*l*omega*sint) / (mc + mp)
	thetaacc := (g*sint - cost*temp) / (l * (4.0/3.0 - mp*cost*cost/(mc+mp)))
	xacc := temp - mp*l*thetaacc*cost/(mc+mp)

	return dynamo.State{vel, xacc, omega, thetaacc}
}
