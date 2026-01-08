package metrics

import (
	"math"

	"github.com/san-kum/dynsim/internal/dynamo"
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

func (e *Energy) Name() string { return e.name }

func (e *Energy) Observe(x dynamo.State, u dynamo.Control, t float64) {
	if len(x) < 2 {
		return
	}
	theta, omega := x[0], x[1]
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

type EnergyDrift struct {
	name          string
	initialEnergy float64
	currentEnergy float64
	maxDrift      float64
	samples       int
	dyn           dynamo.System
}

func NewEnergyDrift(dyn dynamo.System) *EnergyDrift {
	return &EnergyDrift{
		name: "energy_drift",
		dyn:  dyn,
	}
}

func (e *EnergyDrift) Name() string { return e.name }

func (e *EnergyDrift) Observe(x dynamo.State, u dynamo.Control, t float64) {
	ec, ok := e.dyn.(dynamo.Hamiltonian)
	if !ok {
		return
	}

	energy := ec.Energy(x)

	if e.samples == 0 {
		e.initialEnergy = energy
	}

	e.currentEnergy = energy
	e.samples++

	if e.initialEnergy != 0 {
		drift := math.Abs(energy-e.initialEnergy) / math.Abs(e.initialEnergy)
		e.maxDrift = math.Max(e.maxDrift, drift)
	}
}

func (e *EnergyDrift) Value() float64 {
	return e.maxDrift
}

func (e *EnergyDrift) Reset() {
	e.initialEnergy = 0
	e.currentEnergy = 0
	e.maxDrift = 0
	e.samples = 0
}
