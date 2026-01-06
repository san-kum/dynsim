package metrics

import (
	"math"

	"github.com/san-kum/dynsim/internal/sim"
)

type Stability struct {
	name       string
	threshold  float64
	violations int
	samples    int
}

func NewStability(threshold float64) *Stability {
	return &Stability{
		name:      "stability",
		threshold: threshold,
	}
}

func (s *Stability) Name() string {
	return s.name
}

func (s *Stability) Observe(x sim.State, u sim.Control, t float64) {
	s.samples++
	for _, val := range x {
		if math.Abs(val) > s.threshold {
			s.violations++
			break
		}
	}
}

func (s *Stability) Value() float64 {
	if s.samples == 0 {
		return 1.0
	}
	return 1.0 - float64(s.violations)/float64(s.samples)
}

func (s *Stability) Reset() {
	s.violations = 0
	s.samples = 0
}
