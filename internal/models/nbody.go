package models

import (
	"math"

	"github.com/san-kum/dynsim/internal/sim"
)

type NBody struct {
	NumBodies int
	Masses    []float64
	G         float64
}

func NewNBody(n int) *NBody {
	masses := make([]float64, n)
	for i := range masses {
		masses[i] = 1.0
	}
	return &NBody{
		NumBodies: n,
		Masses:    masses,
		G:         1.0,
	}
}

func (nb *NBody) StateDim() int {
	return nb.NumBodies * 4
}

func (nb *NBody) ControlDim() int {
	return 0
}

func (nb *NBody) Derivative(x sim.State, u sim.Control, t float64) sim.State {
	n := nb.NumBodies
	dx := make(sim.State, len(x))

	for i := 0; i < n; i++ {
		dx[i*4] = x[i*4+2]
		dx[i*4+1] = x[i*4+3]

		ax, ay := 0.0, 0.0
		for j := 0; j < n; j++ {
			if i == j {
				continue
			}

			rx := x[j*4] - x[i*4]
			ry := x[j*4+1] - x[i*4+1]
			r := math.Sqrt(rx*rx + ry*ry)

			if r > 1e-6 {
				f := nb.G * nb.Masses[j] / (r * r * r)
				ax += f * rx
				ay += f * ry
			}
		}

		dx[i*4+2] = ax
		dx[i*4+3] = ay
	}

	return dx
}
