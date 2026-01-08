package physics

import "github.com/san-kum/dynsim/internal/dynamo"

type Rossler struct{ a, b, c float64 }

func NewRossler() *Rossler         { return &Rossler{0.2, 0.2, 5.7} }
func (r *Rossler) StateDim() int   { return 3 }
func (r *Rossler) ControlDim() int { return 0 }

// Derive calculates the Rossler attractor derivatives.
func (r *Rossler) Derive(s dynamo.State, _ dynamo.Control, _ float64) dynamo.State {
	return dynamo.State{-s[1] - s[2], s[0] + r.a*s[1], r.b + s[2]*(s[0]-r.c)}
}
func (r *Rossler) DefaultState() dynamo.State { return dynamo.State{1.0, 1.0, 1.0} }
func (r *Rossler) GetParams() map[string]float64 {
	return map[string]float64{"a": r.a, "b": r.b, "c": r.c}
}
func (r *Rossler) SetParam(n string, v float64) {
	switch n {
	case "a":
		r.a = v
	case "b":
		r.b = v
	case "c":
		r.c = v
	}
}
