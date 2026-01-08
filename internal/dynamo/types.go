package dynamo

import (
	"fmt"
	"math"
)

type State []float64

func (s State) Clone() State {
	c := make(State, len(s))
	copy(c, s)
	return c
}

func (s State) IsValid() bool {
	for _, v := range s {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return false
		}
	}
	return true
}

func (s State) Norm() float64 {
	sum := 0.0
	for _, v := range s {
		sum += v * v
	}
	return math.Sqrt(sum)
}

func (s State) Add(other State) State {
	result := make(State, len(s))
	for i := range s {
		if i < len(other) {
			result[i] = s[i] + other[i]
		} else {
			result[i] = s[i]
		}
	}
	return result
}

func (s State) Scale(factor float64) State {
	result := make(State, len(s))
	for i := range s {
		result[i] = s[i] * factor
	}
	return result
}

func (s State) Sub(other State) State {
	result := make(State, len(s))
	for i := range s {
		if i < len(other) {
			result[i] = s[i] - other[i]
		} else {
			result[i] = s[i]
		}
	}
	return result
}

type Control []float64

type System interface {
	Derive(x State, u Control, t float64) State
	StateDim() int
	ControlDim() int
}

type Hamiltonian interface {
	Energy(x State) float64
}

type Integrator interface {
	Step(dyn System, x State, u Control, t float64, dt float64) State
}

type AdaptiveIntegrator interface {
	Integrator
	StepAdaptive(dyn System, x State, u Control, t, dt, tol float64) (State, float64, error)
}

type Controller interface {
	Compute(x State, t float64) Control
}

type Metric interface {
	Name() string
	Observe(x State, u Control, t float64)
	Value() float64
	Reset()
}

type Observer interface {
	OnStep(x State, u Control, t float64)
}

type Configurable interface {
	GetParams() map[string]float64
	SetParam(name string, value float64) error
}

type Config struct {
	Dt            float64
	Duration      float64
	Seed          int64
	Tolerance     float64
	MaxDt         float64
	MinDt         float64
	Adaptive      bool
	ValidateState bool
}

func DefaultConfig() Config {
	return Config{
		Dt:            0.01,
		Duration:      10.0,
		Tolerance:     1e-6,
		MaxDt:         0.1,
		MinDt:         1e-8,
		Adaptive:      false,
		ValidateState: true,
	}
}

type Result struct {
	States      []State
	Controls    []Control
	Times       []float64
	Metrics     map[string]float64
	EnergyDrift float64
	StepsTaken  int
	Errors      []error
}

type SimError struct {
	Time    float64
	Step    int
	Message string
}

func (e SimError) Error() string {
	return fmt.Sprintf("step %d (t=%.4f): %s", e.Step, e.Time, e.Message)
}
