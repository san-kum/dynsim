package sim

type State []float64

func (s State) Clone() State {
	c := make(State, len(s))
	copy(c, s)
	return c
}

type Control []float64

type Dynamics interface {
	Derivative(x State, u Control, t float64) State
	StateDim() int
	ControlDim() int
}

type Integrator interface {
	Step(dyn Dynamics, x State, u Control, t float64, dt float64) State
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

type Config struct {
	Dt       float64
	Duration float64
	Seed     int64
}

type Result struct {
	States   []State
	Controls []Control
	Times    []float64
	Metrics  map[string]float64
}
