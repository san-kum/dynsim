package physics

import (
	"math"

	"github.com/san-kum/dynsim/internal/dynamo"
)

// Gyroscope simulates a spinning top using Euler's equations.
type Gyroscope struct {
	I1, I2, I3, Gravity, Mass, Length float64
}

func NewGyroscope() *Gyroscope {
	return &Gyroscope{1.0, 1.0, 2.0, 9.81, 1.0, 0.5}
}

func (g *Gyroscope) StateDim() int   { return 6 }
func (g *Gyroscope) ControlDim() int { return 0 }

// Derive computes the derivatives for angular velocity and Euler angles.
func (g *Gyroscope) Derive(s dynamo.State, _ dynamo.Control, _ float64) dynamo.State {
	if len(s) < 6 {
		return make(dynamo.State, 6)
	}
	w1, w2, w3, th, _, _ := s[0], s[1], s[2], s[3], s[4], s[5]
	sinT, cosT := math.Sin(th), math.Cos(th)
	if math.Abs(sinT) < 1e-10 {
		sinT = 1e-10
	}
	dW1 := ((g.I2-g.I3)/g.I1)*w2*w3 + (g.Mass*g.Gravity*g.Length*sinT)/g.I1
	dW2, dW3 := ((g.I3-g.I1)/g.I2)*w3*w1, ((g.I1-g.I2)/g.I3)*w1*w2
	return dynamo.State{dW1, dW2, dW3, w1, w2 / sinT, w3 - w2*cosT/sinT}
}

func (g *Gyroscope) DefaultState() dynamo.State {
	return dynamo.State{0.0, 0.0, 10.0, 0.3, 0.0, 0.0}
}

func (g *Gyroscope) Energy(s dynamo.State) float64 {
	if len(s) < 6 {
		return 0
	}
	w1, w2, w3, th := s[0], s[1], s[2], s[3]
	return 0.5*(g.I1*w1*w1+g.I2*w2*w2+g.I3*w3*w3) + g.Mass*g.Gravity*g.Length*math.Cos(th)
}

func (g *Gyroscope) GetParams() map[string]float64 {
	return map[string]float64{"I1": g.I1, "I2": g.I2, "I3": g.I3, "gravity": g.Gravity, "mass": g.Mass, "length": g.Length}
}

func (g *Gyroscope) SetParam(n string, v float64) error {
	switch n {
	case "I1":
		g.I1 = v
	case "I2":
		g.I2 = v
	case "I3":
		g.I3 = v
	case "gravity":
		g.Gravity = v
	case "mass":
		g.Mass = v
	case "length":
		g.Length = v
	}
	return nil
}
