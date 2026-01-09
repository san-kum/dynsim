package physics

import (
	"math"
	"math/rand"

	"github.com/san-kum/dynsim/internal/dynamo"
)

// SPH implements Smoothed Particle Hydrodynamics.
// Simulates fluid flow using particle systems.
type SPH struct {
	N                       int
	H, Rho0, K, Mu, Gravity float64
	Mass                    float64
	BoundsX, BoundsY        float64
}

// NewSPH creates a dam break scenario
func NewSPH(n int) *SPH {
	if n < 100 {
		n = 100
	}
	return &SPH{
		N: n, H: 2.0, Rho0: 1.0, K: 50.0, Mu: 0.1, Gravity: 9.81,
		Mass: 1.0, BoundsX: 60, BoundsY: 40,
	}
}

func (s *SPH) StateDim() int   { return s.N * 4 } // x, y, vx, vy
func (s *SPH) ControlDim() int { return 3 }       // [CursorX, CursorY, Strength]

// kernels because math is hard
func poly6(r2, h2 float64) float64 {
	if r2 > h2 {
		return 0
	}
	return 315.0 / (64.0 * math.Pi * math.Pow(h2, 4.5)) * math.Pow(h2-r2, 3)
}

func spikyGrad(r, h float64) float64 {
	if r > h || r < 1e-6 {
		return 0
	}
	return -45.0 / (math.Pi * math.Pow(h, 6)) * math.Pow(h-r, 2)
}

func viscLap(r, h float64) float64 {
	if r > h {
		return 0
	}
	return 45.0 / (math.Pi * math.Pow(h, 6)) * (h - r)
}

func (s *SPH) Derive(state dynamo.State, u dynamo.Control, _ float64) dynamo.State {
	n, h2 := s.N, s.H*s.H
	deriv := make(dynamo.State, n*4)
	rho, press := make([]float64, n), make([]float64, n)

	// Interaction Force (Hand of God)
	cursorX, cursorY, cursorStr := 0.0, 0.0, 0.0
	if len(u) == 3 {
		cursorX, cursorY, cursorStr = u[0], u[1], u[2]
	}

	// density & pressure
	for i := 0; i < n; i++ {
		rho[i] = 0
		xi, yi := state[i*4], state[i*4+1]
		for j := 0; j < n; j++ {
			dx, dy := xi-state[j*4], yi-state[j*4+1]
			r2 := dx*dx + dy*dy
			if r2 < h2 {
				rho[i] += s.Mass * poly6(r2, h2)
			}
		}
		press[i] = s.K * (rho[i] - s.Rho0)
	}

	// forces
	for i := 0; i < n; i++ {
		fx, fy := 0.0, -s.Gravity*rho[i] // gravity
		xi, yi := state[i*4], state[i*4+1]
		vxi, vyi := state[i*4+2], state[i*4+3]

		for j := 0; j < n; j++ {
			if i == j {
				continue
			}
			dx, dy := xi-state[j*4], yi-state[j*4+1]
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < s.H {
				// pressure force
				fp := -s.Mass * (press[i] + press[j]) / (2 * rho[j]) * spikyGrad(dist, s.H)
				fx += fp * dx / dist
				fy += fp * dy / dist

				// viscosity force
				fv := s.Mu * s.Mass * viscLap(dist, s.H) / rho[j]
				fx += fv * (state[j*4+2] - vxi)
				fy += fv * (state[j*4+3] - vyi)
			}
		}

		// Interaction (Mouse Force)
		if cursorStr != 0 {
			dx, dy := xi-cursorX, yi-cursorY
			dist := math.Sqrt(dx*dx + dy*dy)
			radius := 10.0 // Interaction radius
			if dist < radius {
				// Radial force: F = Str * (1 - dist/R)
				// Repel if Str < 0, Attract if Str > 0
				// Actually, let's vector direction:
				// To attract (Str > 0), we want Force towards Cursor.
				// dx is (Particle - Cursor). So Force should be -dx.
				strength := cursorStr * (1.0 - dist/radius) * 5.0 // Scale factor
				fx -= strength * dx / (dist + 0.1) * rho[i]       // Scale by density to move mass
				fy -= strength * dy / (dist + 0.1) * rho[i]
			}
		}

		// boundary collision (soft repulsion)
		if xi < 0 {
			fx += 500 * -xi
		}
		if xi > s.BoundsX {
			fx -= 500 * (xi - s.BoundsX)
		}
		if yi < 0 {
			fy += 500 * -yi
		}
		if yi > s.BoundsY {
			fy -= 500 * (yi - s.BoundsY)
		}

		deriv[i*4], deriv[i*4+1] = vxi, vyi
		deriv[i*4+2], deriv[i*4+3] = fx/rho[i], fy/rho[i]
	}
	return deriv
}

// dam break setup
func (s *SPH) DefaultState() dynamo.State {
	st := make(dynamo.State, s.N*4)
	cols := int(math.Sqrt(float64(s.N)))
	for i := 0; i < s.N; i++ {
		r, c := i/cols, i%cols
		st[i*4] = float64(c)*s.H*0.5 + 1.0 + rand.Float64()*0.1
		st[i*4+1] = float64(r)*s.H*0.5 + 1.0 + rand.Float64()*0.1
	}
	return st
}

func (s *SPH) GetParams() map[string]float64 {
	return map[string]float64{"h": s.H, "rho0": s.Rho0, "stiffness": s.K, "viscosity": s.Mu, "gravity": s.Gravity}
}

func (s *SPH) SetParam(n string, v float64) error {
	switch n {
	case "h":
		s.H = v
	case "rho0":
		s.Rho0 = v
	case "stiffness":
		s.K = v
	case "viscosity":
		s.Mu = v
	case "gravity":
		s.Gravity = v
	}
	return nil
}
