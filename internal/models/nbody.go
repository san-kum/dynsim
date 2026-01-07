package models

import (
	"math"

	"github.com/san-kum/dynsim/internal/sim"
)

type NBody struct {
	NumBodies int
	Masses    []float64
	G         float64
	Softening float64
	ax, ay    []float64
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
		Softening: 0.01,
		ax:        make([]float64, n),
		ay:        make([]float64, n),
	}
}

func (nb *NBody) StateDim() int   { return nb.NumBodies * 4 }
func (nb *NBody) ControlDim() int { return 0 }

func (nb *NBody) Derivative(x sim.State, u sim.Control, t float64) sim.State {
	n := nb.NumBodies
	dx := make(sim.State, len(x))

	for i := 0; i < n; i++ {
		nb.ax[i] = 0
		nb.ay[i] = 0
	}

	eps2 := nb.Softening * nb.Softening
	for i := 0; i < n; i++ {
		xi, yi := x[i*4], x[i*4+1]

		for j := i + 1; j < n; j++ {
			xj, yj := x[j*4], x[j*4+1]

			rx := xj - xi
			ry := yj - yi
			r2 := rx*rx + ry*ry + eps2

			rInv := 1.0 / math.Sqrt(r2)
			r3Inv := rInv * rInv * rInv

			fij := nb.G * nb.Masses[j] * r3Inv
			nb.ax[i] += fij * rx
			nb.ay[i] += fij * ry

			fji := nb.G * nb.Masses[i] * r3Inv
			nb.ax[j] -= fji * rx
			nb.ay[j] -= fji * ry
		}
	}

	for i := 0; i < n; i++ {
		dx[i*4] = x[i*4+2]
		dx[i*4+1] = x[i*4+3]
		dx[i*4+2] = nb.ax[i]
		dx[i*4+3] = nb.ay[i]
	}

	return dx
}

func (nb *NBody) Energy(x sim.State) float64 {
	n := nb.NumBodies
	ke := 0.0
	pe := 0.0

	for i := 0; i < n; i++ {
		vx, vy := x[i*4+2], x[i*4+3]
		ke += 0.5 * nb.Masses[i] * (vx*vx + vy*vy)

		for j := i + 1; j < n; j++ {
			rx := x[j*4] - x[i*4]
			ry := x[j*4+1] - x[i*4+1]
			r := math.Sqrt(rx*rx + ry*ry + nb.Softening*nb.Softening)
			pe -= nb.G * nb.Masses[i] * nb.Masses[j] / r
		}
	}

	return ke + pe
}

func (nb *NBody) Momentum(x sim.State) (px, py float64) {
	for i := 0; i < nb.NumBodies; i++ {
		px += nb.Masses[i] * x[i*4+2]
		py += nb.Masses[i] * x[i*4+3]
	}
	return
}

func (nb *NBody) AngularMomentum(x sim.State) float64 {
	L := 0.0
	for i := 0; i < nb.NumBodies; i++ {
		xi, yi := x[i*4], x[i*4+1]
		vx, vy := x[i*4+2], x[i*4+3]
		L += nb.Masses[i] * (xi*vy - yi*vx)
	}
	return L
}
