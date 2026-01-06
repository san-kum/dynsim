package metrics

import (
	"math"

	"github.com/san-kum/dynsim/internal/sim"
)

type Energy struct {
	name        string
	mass        float64
	length      float64
	gravity     float64
	samples     int
	totalEnergy float64
}

func NewEnergy(mass, length, gravity float64) *Energy {
	return &Energy{
		name:    "energy",
		mass:    mass,
		length:  length,
		gravity: gravity,
	}
}

func (e *Energy) Name() string {
	return e.name
}

func (e *Energy) Observe(x sim.State, u sim.Control, t float64) {
	if len(x) < 2 {
		return
	}

	theta := x[0]
	omega := x[1]

	ke := 0.5 * e.mass * e.length * e.length * omega * omega
	pe := e.mass * e.gravity * e.length * (1 - math.Cos(theta))

	e.totalEnergy += ke + pe
	e.samples++
}

func (e *Energy) Value() float64 {
	if e.samples == 0 {
		return 0
	}
	return e.totalEnergy / float64(e.samples)
}

func (e *Energy) Reset() {
	e.totalEnergy = 0
	e.samples = 0
}
