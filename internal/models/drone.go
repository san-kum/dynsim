package models

import (
	"math"

	"github.com/san-kum/dynsim/internal/sim"
)

type Drone struct {
	Mass, Inertia, ArmLength float64
	Gravity, DragCoeff       float64
	AngDrag                  float64
}

func NewDrone() *Drone {
	return &Drone{
		Mass:      DefaultMass,
		Inertia:   0.1,
		ArmLength: 0.25,
		Gravity:   DefaultGravity,
		DragCoeff: 0.1,
		AngDrag:   0.05,
	}
}

func (d *Drone) StateDim() int   { return 6 }
func (d *Drone) ControlDim() int { return 2 }

func (d *Drone) Derivative(x sim.State, u sim.Control, t float64) sim.State {
	theta, vx, vy, omega := x[2], x[3], x[4], x[5]

	thrustL, thrustR := 0.0, 0.0
	if len(u) >= 2 {
		thrustL, thrustR = u[0], u[1]
	} else if len(u) >= 1 {
		thrustL, thrustR = u[0]/2, u[0]/2
	}

	thrustL = math.Max(0, thrustL)
	thrustR = math.Max(0, thrustR)

	totalThrust := thrustL + thrustR
	torque := (thrustR - thrustL) * d.ArmLength

	sin, cos := math.Sin(theta), math.Cos(theta)
	fx := -totalThrust*sin - d.DragCoeff*vx
	fy := totalThrust*cos - d.Mass*d.Gravity - d.DragCoeff*vy

	ax := fx / d.Mass
	ay := fy / d.Mass
	alpha := (torque - d.AngDrag*omega) / d.Inertia

	return sim.State{vx, vy, omega, ax, ay, alpha}
}

func (d *Drone) HoverThrust() float64 {
	return d.Mass * d.Gravity / 2.0
}

func (d *Drone) Energy(x sim.State) float64 {
	y, vx, vy, omega := x[1], x[3], x[4], x[5]
	ke := 0.5 * d.Mass * (vx*vx + vy*vy)
	keRot := 0.5 * d.Inertia * omega * omega
	pe := d.Mass * d.Gravity * y
	return ke + keRot + pe
}
