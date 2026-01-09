package physics

import (
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"sync"

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
	// Realistic Galaxy Generator (Bulge + Spiral Disk + Halo)
	n := nb.NumBodies
	state := make(dynamo.State, n*4)
	rnd := rand.New(rand.NewSource(42)) // Fixed seed for reproducibility

	// Distribution Ratios
	nBulge := int(float64(n) * 0.10)
	nHalo := int(float64(n) * 0.05)
	nDisk := n - nBulge - nHalo

	// 1. Generate Positions
	// Central Black Hole (Body 0) - Supermassive anchor
	nb.Masses[0] = 500000.0 // Supermassive (Scaled for stability)
	state[0], state[1] = 0, 0
	state[2], state[3] = 0, 0

	idx := 1

	// Bulge (Spherical, dense, hot)
	for i := 0; i < nBulge; i++ {
		if idx >= n {
			break
		}
		r := math.Abs(rnd.NormFloat64()) * 20.0 // Compact
		theta := rnd.Float64() * 2 * math.Pi
		state[idx*4] = r * math.Cos(theta)
		state[idx*4+1] = r * math.Sin(theta)
		idx++
	}

	// Disk (Spiral Arms, Exponential Decay)
	arms := 2.0
	armTwist := 5.0
	for i := 0; i < nDisk; i++ {
		if idx >= n {
			break
		}
		// Exponential falloff for density
		r := 20.0 + math.Abs(rnd.NormFloat64())*100.0 + rnd.ExpFloat64()*30.0
		if r > 300 {
			r = 300
		}

		// Spiral Angle
		baseAngle := (float64(i%int(arms)) / arms) * 2 * math.Pi
		angle := baseAngle + armTwist*math.Log(r/20.0) + (rnd.Float64()-0.5)*0.5

		state[idx*4] = r * math.Cos(angle)
		state[idx*4+1] = r * math.Sin(angle)
		idx++
	}

	// Halo (Distant, Scattered)
	for i := 0; i < nHalo; i++ {
		if idx >= n {
			break
		}
		r := 100.0 + math.Abs(rnd.NormFloat64())*200.0
		theta := rnd.Float64() * 2 * math.Pi
		state[idx*4] = r * math.Cos(theta)
		state[idx*4+1] = r * math.Sin(theta)
		idx++
	}

	// 2. Stability Pre-Pass (The "Cold Start" Fix)
	// Calculate exact gravitational acceleration for every particle based on the generated configuration using the Parallel Solver.
	// Then set velocity to circular orbit speed.

	fmt.Println("Simulating initial gravity for stability...")
	ax, ay := nb.computeForcesCPU(state)

	for i := 1; i < n; i++ {
		xi, yi := state[i*4], state[i*4+1]
		dist := math.Sqrt(xi*xi + yi*yi)
		if dist < 0.1 {
			continue
		}

		// Net Acceleration Vector
		aci_x, aci_y := ax[i], ay[i]

		// Radial acceleration magnitude
		a_mag := math.Sqrt(aci_x*aci_x + aci_y*aci_y)

		// Circular Velocity v = sqrt(a * r)
		v := math.Sqrt(a_mag * dist)

		// Radial unit vector:
		ux, uy := xi/dist, yi/dist

		// Direction: Tangent to radius (-y, x) for CCW rotation
		state[i*4+2] = -v * uy
		state[i*4+3] = v * ux

		// Add "Temperature" (Random Velocity Dispersion)
		dispersion := 0.0
		if i < nBulge {
			dispersion = v * 0.4
		} else if i > n-nHalo {
			dispersion = v * 0.5
		} else {
			dispersion = v * 0.05
		}

		state[i*4+2] += (rnd.Float64() - 0.5) * dispersion
		state[i*4+3] += (rnd.Float64() - 0.5) * dispersion
	}

	return state
}

func (nb *NBody) StateDim() int   { return nb.NumBodies * 4 }
func (nb *NBody) ControlDim() int { return 3 } // [CursorX, CursorY, Strength]

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

	// Interaction (Hand of God)
	cursorX, cursorY, cursorStr := 0.0, 0.0, 0.0
	if len(u) == 3 {
		cursorX, cursorY, cursorStr = u[0], u[1], u[2]
	}

	for i := 0; i < n; i++ {
		dx[i*4] = x[i*4+2]
		dx[i*4+1] = x[i*4+3]

		// Add Interaction Force
		ix, iy := 0.0, 0.0
		if cursorStr != 0 {
			xi, yi := x[i*4], x[i*4+1]
			rx := cursorX - xi
			ry := cursorY - yi
			dist2 := rx*rx + ry*ry + 5.0 // Softening for cursor
			dist := math.Sqrt(dist2)
			f := cursorStr * 20.0 / (dist * dist2) // InvSq Force

			ix = f * rx
			iy = f * ry
		}

		dx[i*4+2] = ax[i] + ix
		dx[i*4+3] = ay[i] + iy
	}

	return dx
}

func (nb *NBody) computeForcesCPU(x dynamo.State) ([]float64, []float64) {
	n := nb.NumBodies
	ax := make([]float64, n)
	ay := make([]float64, n)
	eps2 := nb.Softening * nb.Softening

	// Parallel Execution
	numWorkers := runtime.NumCPU()
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	chunkSize := (n + numWorkers - 1) / numWorkers

	for w := 0; w < numWorkers; w++ {
		start := w * chunkSize
		end := start + chunkSize
		if end > n {
			end = n
		}

		go func(start, end int) {
			defer wg.Done()
			for i := start; i < end; i++ {
				xi, yi := x[i*4], x[i*4+1]
				fx, fy := 0.0, 0.0

				for j := 0; j < n; j++ {
					if i == j {
						continue
					}
					xj, yj := x[j*4], x[j*4+1]
					rx := xj - xi
					ry := yj - yi
					dist2 := rx*rx + ry*ry + eps2
					invDist := 1.0 / math.Sqrt(dist2)
					invDist3 := invDist * invDist * invDist
					f := nb.G * nb.Masses[j] * invDist3
					fx += f * rx
					fy += f * ry
				}
				ax[i] = fx
				ay[i] = fy
			}
		}(start, end)
	}

	wg.Wait()
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