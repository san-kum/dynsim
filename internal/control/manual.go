package control

import "github.com/san-kum/dynsim/internal/dynamo"

// ManualController passes a manually set control vector to the system.
// Used for "Hand of God" user interaction (Mouse Force).
type ManualController struct {
	U []float64 // The current control vector [x, y, strength]
}

func NewManual() *ManualController {
	return &ManualController{
		U: make([]float64, 3), // [TargetX, TargetY, Key/Strength]
	}
}

// SetControl updates the control vector.
func (c *ManualController) SetControl(u []float64) {
	if len(u) != 3 {
		return
	}
	copy(c.U, u)
}

// Compute returns the stored control vector.
func (c *ManualController) Compute(state dynamo.State, t float64) dynamo.Control {
	return c.U
}
