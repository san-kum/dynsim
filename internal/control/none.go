package control

import "github.com/san-kum/dynsim/internal/dynamo"

type None struct {
	dim int
}

func NewNone(dim int) *None {
	return &None{
		dim: dim,
	}
}

func (n *None) Compute(x dynamo.State, t float64) dynamo.Control {
	return make(dynamo.Control, n.dim)
}
