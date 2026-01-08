package physics

import (
	"math"

	"github.com/san-kum/dynsim/internal/compute"
	"github.com/san-kum/dynsim/internal/dynamo"
)

type NBody struct {
	NumBodies int
	Masses    []float64
	G         float64
	Softening float64
	UseGPU    bool
	positions []float64
}

// NewNBody creates and returns an NBody configured for n bodies.
// 
// The returned NBody has masses initialized to 1.0 for every body, gravitational
// constant G set to 1.0, softening length set to 0.01, the UseGPU flag set
// according to the compute backend availability, and an internal positions
// buffer sized for n (x,y) pairs.
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
		UseGPU:    compute.GetBackend().Available(),
		positions: make([]float64, n*2),
	}
}

func (nb *NBody) DefaultState() dynamo.State {
	n := nb.NumBodies
	state := make(dynamo.State, n*4)
	for i := 0; i < n; i++ {
		angle := float64(i) * 2.0 * 3.14159 / float64(n)
		radius := 1.0
		state[i*4] = radius * math.Cos(angle)
		state[i*4+1] = radius * math.Sin(angle)
		state[i*4+2] = -math.Sin(angle) * 0.5
		state[i*4+3] = math.Cos(angle) * 0.5
	}
	return state
}

func (nb *NBody) StateDim() int   { return nb.NumBodies * 4 }
func (nb *NBody) ControlDim() int { return 0 }

func (nb *NBody) Derive(x dynamo.State, u dynamo.Control, t float64) dynamo.State {
	n := nb.NumBodies
	dx := make(dynamo.State, len(x))

	for i := 0; i < n; i++ {
		nb.positions[i*2] = x[i*4]
		nb.positions[i*2+1] = x[i*4+1]
	}

	var ax, ay []float64

	if nb.UseGPU && n >= 32 {
		ax, ay = compute.GetBackend().NBodyForces(nb.positions, nb.Masses, nb.G, nb.Softening)
	} else {
		ax, ay = nb.computeForcesCPU(x)
	}

	for i := 0; i < n; i++ {
		dx[i*4] = x[i*4+2]
		dx[i*4+1] = x[i*4+3]
		dx[i*4+2] = ax[i]
		dx[i*4+3] = ay[i]
	}

	return dx
}

func (nb *NBody) computeForcesCPU(x dynamo.State) ([]float64, []float64) {
	n := nb.NumBodies
	ax := make([]float64, n)
	ay := make([]float64, n)
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
			ax[i] += fij * rx
			ay[i] += fij * ry

			fji := nb.G * nb.Masses[i] * r3Inv
			ax[j] -= fji * rx
			ay[j] -= fji * ry
		}
	}

	return ax, ay
}

func (nb *NBody) Energy(x dynamo.State) float64 {
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

func (nb *NBody) Momentum(x dynamo.State) (px, py float64) {
	for i := 0; i < nb.NumBodies; i++ {
		px += nb.Masses[i] * x[i*4+2]
		py += nb.Masses[i] * x[i*4+3]
	}
	return
}

func (nb *NBody) AngularMomentum(x dynamo.State) float64 {
	L := 0.0
	for i := 0; i < nb.NumBodies; i++ {
		xi, yi := x[i*4], x[i*4+1]
		vx, vy := x[i*4+2], x[i*4+3]
		L += nb.Masses[i] * (xi*vy - yi*vx)
	}
	return L
}