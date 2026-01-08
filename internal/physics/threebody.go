package physics

import (
	"math"

	"github.com/san-kum/dynsim/internal/dynamo"
)

// ThreeBody implements a gravitational three-body problem.
// State: [x1, y1, vx1, vy1, x2, y2, vx2, vy2, x3, y3, vx3, vy3]
// Each body has mass, position (x, y), and velocity (vx, vy).
type ThreeBody struct {
	m1, m2, m3 float64 // Masses
	g          float64 // Gravitational constant
	softening  float64 // Prevent singularities
}

func NewThreeBody() *ThreeBody {
	return &ThreeBody{
		m1:        1.0,
		m2:        1.0,
		m3:        1.0,
		g:         1.0,
		softening: 0.1,
	}
}

func (t *ThreeBody) StateDim() int   { return 12 }
func (t *ThreeBody) ControlDim() int { return 0 }

func (t *ThreeBody) Derive(state dynamo.State, _ dynamo.Control, _ float64) dynamo.State {
	x1, y1, vx1, vy1 := state[0], state[1], state[2], state[3]
	x2, y2, vx2, vy2 := state[4], state[5], state[6], state[7]
	x3, y3, vx3, vy3 := state[8], state[9], state[10], state[11]

	// Distance calculations with softening
	r12 := math.Sqrt((x2-x1)*(x2-x1) + (y2-y1)*(y2-y1) + t.softening*t.softening)
	r13 := math.Sqrt((x3-x1)*(x3-x1) + (y3-y1)*(y3-y1) + t.softening*t.softening)
	r23 := math.Sqrt((x3-x2)*(x3-x2) + (y3-y2)*(y3-y2) + t.softening*t.softening)

	// Accelerations on body 1
	ax1 := t.g * t.m2 * (x2 - x1) / (r12 * r12 * r12)
	ax1 += t.g * t.m3 * (x3 - x1) / (r13 * r13 * r13)
	ay1 := t.g * t.m2 * (y2 - y1) / (r12 * r12 * r12)
	ay1 += t.g * t.m3 * (y3 - y1) / (r13 * r13 * r13)

	// Accelerations on body 2
	ax2 := t.g * t.m1 * (x1 - x2) / (r12 * r12 * r12)
	ax2 += t.g * t.m3 * (x3 - x2) / (r23 * r23 * r23)
	ay2 := t.g * t.m1 * (y1 - y2) / (r12 * r12 * r12)
	ay2 += t.g * t.m3 * (y3 - y2) / (r23 * r23 * r23)

	// Accelerations on body 3
	ax3 := t.g * t.m1 * (x1 - x3) / (r13 * r13 * r13)
	ax3 += t.g * t.m2 * (x2 - x3) / (r23 * r23 * r23)
	ay3 := t.g * t.m1 * (y1 - y3) / (r13 * r13 * r13)
	ay3 += t.g * t.m2 * (y2 - y3) / (r23 * r23 * r23)

	return dynamo.State{
		vx1, vy1, ax1, ay1,
		vx2, vy2, ax2, ay2,
		vx3, vy3, ax3, ay3,
	}
}

func (t *ThreeBody) DefaultState() dynamo.State {
	// Figure-8 solution initial conditions (approximately)
	return dynamo.State{
		-1.0, 0.0, 0.347, 0.532, // Body 1
		1.0, 0.0, 0.347, 0.532, // Body 2
		0.0, 0.0, -0.694, -1.064, // Body 3
	}
}

// GetParams implements dynamo.Configurable
func (t *ThreeBody) GetParams() map[string]float64 {
	return map[string]float64{
		"m1": t.m1,
		"m2": t.m2,
		"m3": t.m3,
		"g":  t.g,
	}
}

// SetParam implements dynamo.Configurable
func (t *ThreeBody) SetParam(name string, value float64) {
	switch name {
	case "m1":
		t.m1 = value
	case "m2":
		t.m2 = value
	case "m3":
		t.m3 = value
	case "g":
		t.g = value
	}
}
