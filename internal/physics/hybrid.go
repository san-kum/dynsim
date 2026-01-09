package physics

import (
	"math"
	"math/rand"

	"github.com/san-kum/dynsim/internal/compute"
	"github.com/san-kum/dynsim/internal/dynamo"
)

type Hybrid struct {
	Stars *NBody
	Gas   *SPH

	// Combined Buffers for GPU Gravity
	allPositions []float64
	allMasses    []float64
}

func NewHybrid(nStars, nGas int) *Hybrid {
	return &Hybrid{
		Stars:        NewNBody(nStars),
		Gas:          NewSPH(nGas),
		allPositions: make([]float64, (nStars+nGas)*2),
		allMasses:    make([]float64, nStars+nGas),
	}
}

func (h *Hybrid) StateDim() int {
	return h.Stars.StateDim() + h.Gas.StateDim()
}

func (h *Hybrid) ControlDim() int {
	return 3
}

func (h *Hybrid) DefaultState() dynamo.State {
	// Initialize Stars using Galaxy Gen
	starState := h.Stars.DefaultState()

	// Initialize Gas (Disk + Random Clouds)
	gasState := make(dynamo.State, h.Gas.N*4)
	rnd := rand.New(rand.NewSource(1337))

	nGas := h.Gas.N
	for i := 0; i < nGas; i++ {
		// Gas roughly follows the spiral arms but is more diffuse
		r := 20.0 + math.Abs(rnd.NormFloat64())*100.0 + rnd.ExpFloat64()*30.0
		if r > 300 {
			r = 300
		}

		angle := rnd.Float64() * 2 * math.Pi

		// Position
		gasState[i*4] = r * math.Cos(angle)
		gasState[i*4+1] = r * math.Sin(angle)

		// Velocity (Orbital)
		// Approximate mass of core = 500000 (Matched with NBody)
		v := math.Sqrt(1.0 * 500000.0 / r)
		gasState[i*4+2] = -v * math.Sin(angle)
		gasState[i*4+3] = v * math.Cos(angle)
	}

	// Combine
	fullState := make(dynamo.State, 0, len(starState)+len(gasState))
	fullState = append(fullState, starState...)
	fullState = append(fullState, gasState...)

	return fullState
}

// Spatial Hash Grid
const (
	HashP1   = 73856093
	HashP2   = 19349663
	GridSize = 2.0 // Match Smoothing Radius H
)

func hashPos(x, y float64) int {
	ix := int(math.Floor(x / GridSize))
	iy := int(math.Floor(y / GridSize))
	return (ix * HashP1) ^ (iy * HashP2)
}

func (h *Hybrid) Derive(x dynamo.State, u dynamo.Control, t float64) dynamo.State {
	nStars := h.Stars.NumBodies
	nGas := h.Gas.N
	dt := make(dynamo.State, len(x))

	// Split State
	starState := x[:nStars*4]
	gasState := x[nStars*4:]

	// 1. Prepare for Gravity (GPU)
	// Update Combined Buffers
	// Mass is static usually, but let's refresh to be safe or just init once?
	// Optimizing: Only doing positions reduces PCI transfer if backend supports partial updates.
	// But our backend API takes flat arrays.
	for i := 0; i < nStars; i++ {
		h.allMasses[i] = h.Stars.Masses[i]
		h.allPositions[i*2] = starState[i*4]
		h.allPositions[i*2+1] = starState[i*4+1]
	}
	for i := 0; i < nGas; i++ {
		h.allMasses[nStars+i] = h.Gas.Mass
		h.allPositions[(nStars+i)*2] = gasState[i*4]
		h.allPositions[(nStars+i)*2+1] = gasState[i*4+1]
	}

	// 2a. Audio Reactivity (Gravity Pulse)
	// Control vector u: [curX, curY, curStr, Bass, Mid, High]
	baseG := h.Stars.G
	currentG := baseG

	if len(u) >= 4 {
		bass := u[3]
		// Bass pumps gravity: On beat, gravity increases (contraction), off beat, it relaxes
		// "Pulse": G increases by 50%
		currentG = baseG * (1.0 + bass*2.0)
	}

	// Compute Gravity (All-to-All on GPU)
	gx, gy := compute.GetBackend().NBodyForces(h.allPositions, h.allMasses, currentG, h.Stars.Softening)

	// 2. Spatial Grid for SPH
	grid := make(map[int][]int)

	// Populate Grid
	for i := 0; i < nGas; i++ {
		key := hashPos(gasState[i*4], gasState[i*4+1])
		grid[key] = append(grid[key], i)
	}

	// SPH: Density & Pressure
	rho := make([]float64, nGas)
	press := make([]float64, nGas)
	h2 := h.Gas.H * h.Gas.H

	for i := 0; i < nGas; i++ {
		xi, yi := gasState[i*4], gasState[i*4+1]
		rho[i] = 0

		// Neighbor Search (3x3 blocks)
		ix := int(math.Floor(xi / GridSize))
		iy := int(math.Floor(yi / GridSize))

		for xx := ix - 1; xx <= ix+1; xx++ {
			for yy := iy - 1; yy <= iy+1; yy++ {
				key := (xx * HashP1) ^ (yy * HashP2)
				neighbors := grid[key]
				for _, j := range neighbors {
					dx, dy := xi-gasState[j*4], yi-gasState[j*4+1]
					r2 := dx*dx + dy*dy
					if r2 < h2 {
						rho[i] += h.Gas.Mass * poly6(r2, h2)
					}
				}
			}
		}
		press[i] = h.Gas.K * (rho[i] - h.Gas.Rho0)
	}

	// SPH: Forces
	sphFx := make([]float64, nGas)
	sphFy := make([]float64, nGas)

	for i := 0; i < nGas; i++ {
		xi, yi := gasState[i*4], gasState[i*4+1]
		vxi, vyi := gasState[i*4+2], gasState[i*4+3]

		ix := int(math.Floor(xi / GridSize))
		iy := int(math.Floor(yi / GridSize))

		for xx := ix - 1; xx <= ix+1; xx++ {
			for yy := iy - 1; yy <= iy+1; yy++ {
				key := (xx * HashP1) ^ (yy * HashP2)
				neighbors := grid[key]
				for _, j := range neighbors {
					if i == j {
						continue
					}
					dx, dy := xi-gasState[j*4], yi-gasState[j*4+1]
					dist := math.Sqrt(dx*dx + dy*dy)

					if dist < h.Gas.H {
						// Pressure
						fp := -h.Gas.Mass * (press[i] + press[j]) / (2 * rho[j]) * spikyGrad(dist, h.Gas.H)
						sphFx[i] += fp * dx / dist
						sphFy[i] += fp * dy / dist

						// Viscosity
						fv := h.Gas.Mu * h.Gas.Mass * viscLap(dist, h.Gas.H) / rho[j]
						sphFx[i] += fv * (gasState[j*4+2] - vxi)
						sphFy[i] += fv * (gasState[j*4+3] - vyi)
					}
				}
			}
		}
	}

	// 3. Interaction
	cursorX, cursorY, cursorStr := 0.0, 0.0, 0.0
	if len(u) >= 3 {
		cursorX, cursorY, cursorStr = u[0], u[1], u[2]
	}

	// 4. Integration

	// Stars
	for i := 0; i < nStars; i++ {
		dt[i*4] = starState[i*4+2]
		dt[i*4+1] = starState[i*4+3]

		fx, fy := gx[i], gy[i]

		if cursorStr != 0 {
			xi, yi := starState[i*4], starState[i*4+1]
			rx, ry := cursorX-xi, cursorY-yi
			dist2 := rx*rx + ry*ry + 5.0
			dist := math.Sqrt(dist2)
			f := cursorStr * 20.0 / (dist * dist2)
			fx += f * rx
			fy += f * ry
		}

		dt[i*4+2] = fx
		dt[i*4+3] = fy
	}

	// Gas
	for i := 0; i < nGas; i++ {
		idx := nStars*4 + i*4

		dt[idx] = gasState[i*4+2]
		dt[idx+1] = gasState[i*4+3]

		// Forces: Gravity + SPH
		fx := gx[nStars+i] + sphFx[i]/rho[i]
		fy := gy[nStars+i] + sphFy[i]/rho[i]

		if cursorStr != 0 {
			xi, yi := gasState[i*4], gasState[i*4+1]
			rx, ry := cursorX-xi, cursorY-yi
			dist2 := rx*rx + ry*ry + 5.0
			dist := math.Sqrt(dist2)
			f := cursorStr * 20.0 / (dist * dist2)
			fx += f * rx
			fy += f * ry
		}

		// Bounds
		xi := gasState[i*4]
		yi := gasState[i*4+1]
		if xi < -300 {
			fx += 100
		}
		if xi > 300 {
			fx -= 100
		}
		if yi < -300 {
			fy += 100
		}
		if yi > 300 {
			fy -= 100
		}

		dt[idx+2] = fx
		dt[idx+3] = fy
	}

	return dt
}
