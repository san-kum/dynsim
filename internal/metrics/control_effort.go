package metrics

import (
	"math"

	"github.com/san-kum/dynsim/internal/dynamo"
)

type ControlEffort struct {
	name    string
	sum     float64
	samples int
}

func NewControlEffort() *ControlEffort {
	return &ControlEffort{
		name: "control_effort",
	}
}

func (c *ControlEffort) Name() string {
	return c.name
}

func (c *ControlEffort) Observe(x dynamo.State, u dynamo.Control, t float64) {
	for _, val := range u {
		c.sum += math.Abs(val)
	}
	c.samples++
}

func (c *ControlEffort) Value() float64 {
	if c.samples == 0 {
		return 0
	}
	return c.sum / float64(c.samples)
}

func (c *ControlEffort) Reset() {
	c.sum = 0
	c.samples = 0
}
