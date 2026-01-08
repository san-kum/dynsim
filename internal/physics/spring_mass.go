package physics

import "github.com/san-kum/dynsim/internal/dynamo"

const (
	DefaultStiffness = 10.0
	DefaultDamping   = 0.5
)

type SpringMass struct {
	NumMasses int
	Masses    []float64
	Stiffness []float64
	Damping   []float64
}

func NewSpringMass() *SpringMass {
	return &SpringMass{
		NumMasses: 1,
		Masses:    []float64{DefaultMass},
		Stiffness: []float64{DefaultStiffness},
		Damping:   []float64{DefaultDamping},
	}
}

func NewSpringMassChain(n int) *SpringMass {
	masses := make([]float64, n)
	stiffness := make([]float64, n+1)
	damping := make([]float64, n)

	for i := 0; i < n; i++ {
		masses[i] = DefaultMass
		stiffness[i] = DefaultStiffness
		damping[i] = 0.2
	}
	stiffness[n] = DefaultStiffness

	return &SpringMass{
		NumMasses: n,
		Masses:    masses,
		Stiffness: stiffness,
		Damping:   damping,
	}
}

func (s *SpringMass) StateDim() int   { return s.NumMasses * 2 }
func (s *SpringMass) ControlDim() int { return 1 }

func (s *SpringMass) Derive(x dynamo.State, u dynamo.Control, t float64) dynamo.State {
	n := s.NumMasses
	dx := make(dynamo.State, n*2)

	for i := 0; i < n; i++ {
		dx[i] = x[n+i]
	}

	extForce := 0.0
	if len(u) > 0 {
		extForce = u[0]
	}

	for i := 0; i < n; i++ {
		pos, vel := x[i], x[n+i]

		var forceLeft, forceRight float64
		if i == 0 {
			forceLeft = -s.Stiffness[0] * pos
		} else {
			forceLeft = -s.Stiffness[i] * (pos - x[i-1])
		}

		if i == n-1 {
			if len(s.Stiffness) > n {
				forceRight = -s.Stiffness[n] * pos
			}
		} else {
			forceRight = -s.Stiffness[i+1] * (pos - x[i+1])
		}

		totalForce := forceLeft + forceRight - s.Damping[i]*vel
		if i == 0 {
			totalForce += extForce
		}
		dx[n+i] = totalForce / s.Masses[i]
	}

	return dx
}

func (s *SpringMass) Energy(x dynamo.State) float64 {
	n := s.NumMasses
	energy := 0.0

	for i := 0; i < n; i++ {
		v := x[n+i]
		energy += 0.5 * s.Masses[i] * v * v
	}

	for i := 0; i < n; i++ {
		pos := x[i]
		if i == 0 {
			energy += 0.5 * s.Stiffness[0] * pos * pos
		} else {
			stretch := pos - x[i-1]
			energy += 0.5 * s.Stiffness[i] * stretch * stretch
		}
	}

	if len(s.Stiffness) > n {
		energy += 0.5 * s.Stiffness[n] * x[n-1] * x[n-1]
	}

	return energy
}
