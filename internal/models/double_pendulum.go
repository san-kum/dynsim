package models

import (
	"math"

	"github.com/san-kum/dynsim/internal/sim"
)

const (
	DefaultMass    = 1.0
	DefaultLength  = 1.0
	DefaultGravity = 9.81
)

type DoublePendulum struct {
	M1, M2  float64
	L1, L2  float64
	Gravity float64
}

func NewDoublePendulum() *DoublePendulum {
	return &DoublePendulum{
		M1: DefaultMass, M2: DefaultMass,
		L1: DefaultLength, L2: DefaultLength,
		Gravity: DefaultGravity,
	}
}

func (d *DoublePendulum) StateDim() int   { return 4 }
func (d *DoublePendulum) ControlDim() int { return 1 }

func (d *DoublePendulum) Derivative(x sim.State, u sim.Control, t float64) sim.State {
	theta1, theta2, omega1, omega2 := x[0], x[1], x[2], x[3]
	m1, m2, l1, l2, g := d.M1, d.M2, d.L1, d.L2, d.Gravity

	delta := theta2 - theta1
	sinD, cosD := math.Sin(delta), math.Cos(delta)

	tau := 0.0
	if len(u) > 0 {
		tau = u[0]
	}

	den1 := (m1+m2)*l1 - m2*l1*cosD*cosD
	den2 := (l2 / l1) * den1

	alpha1 := (m2*l1*omega1*omega1*sinD*cosD +
		m2*g*math.Sin(theta2)*cosD +
		m2*l2*omega2*omega2*sinD -
		(m1+m2)*g*math.Sin(theta1) + tau) / den1

	alpha2 := (-m2*l2*omega2*omega2*sinD*cosD +
		(m1+m2)*g*math.Sin(theta1)*cosD -
		(m1+m2)*l1*omega1*omega1*sinD -
		(m1+m2)*g*math.Sin(theta2)) / den2

	return sim.State{omega1, omega2, alpha1, alpha2}
}

func (d *DoublePendulum) Energy(x sim.State) float64 {
	theta1, theta2, omega1, omega2 := x[0], x[1], x[2], x[3]
	m1, m2, l1, l2, g := d.M1, d.M2, d.L1, d.L2, d.Gravity

	v1sq := l1 * l1 * omega1 * omega1
	v2sq := l1*l1*omega1*omega1 + l2*l2*omega2*omega2 +
		2*l1*l2*omega1*omega2*math.Cos(theta1-theta2)

	ke := 0.5*m1*v1sq + 0.5*m2*v2sq
	y1 := -l1 * math.Cos(theta1)
	y2 := y1 - l2*math.Cos(theta2)
	pe := m1*g*y1 + m2*g*y2

	return ke + pe
}
