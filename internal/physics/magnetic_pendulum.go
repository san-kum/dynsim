package physics

import (
	"math"

	"github.com/san-kum/dynsim/internal/dynamo"
)

// MagneticPendulum models a pendulum moving over magnets.
type MagneticPendulum struct {
	Magnets                               []Magnet
	Height, Damping, Gravity, MagnetPower float64
}

type Magnet struct {
	X, Y, Strength float64
}

func NewMagneticPendulum() *MagneticPendulum {
	r := 1.5
	m := []Magnet{
		{r * math.Cos(0), r * math.Sin(0), 1.0},
		{r * math.Cos(2*math.Pi/3), r * math.Sin(2*math.Pi/3), 1.0},
		{r * math.Cos(4*math.Pi/3), r * math.Sin(4*math.Pi/3), 1.0},
	}
	return &MagneticPendulum{m, 0.5, 0.2, 0.5, 3.0}
}

func (m *MagneticPendulum) StateDim() int   { return 4 }
func (m *MagneticPendulum) ControlDim() int { return 0 }

func (m *MagneticPendulum) Derive(s dynamo.State, _ dynamo.Control, _ float64) dynamo.State {
	if len(s) < 4 {
		return make(dynamo.State, 4)
	}
	x, y, vx, vy := s[0], s[1], s[2], s[3]
	fx, fy := -m.Gravity*x-m.Damping*vx, -m.Gravity*y-m.Damping*vy
	for _, mag := range m.Magnets {
		dx, dy := mag.X-x, mag.Y-y
		dist := math.Sqrt(dx*dx + dy*dy + m.Height*m.Height)
		if dist < 0.1 {
			dist = 0.1
		}
		f := mag.Strength / math.Pow(dist, m.MagnetPower)
		hd := math.Sqrt(dx*dx + dy*dy)
		if hd > 1e-10 {
			fx += f * dx / hd
			fy += f * dy / hd
		}
	}
	return dynamo.State{vx, vy, fx, fy}
}

func (m *MagneticPendulum) DefaultState() dynamo.State { return dynamo.State{0.5, 0.3, 0, 0} }

func (m *MagneticPendulum) Energy(s dynamo.State) float64 {
	if len(s) < 4 {
		return 0
	}
	x, y, vx, vy := s[0], s[1], s[2], s[3]
	pe, magPE := 0.5*m.Gravity*(x*x+y*y), 0.0
	for _, mag := range m.Magnets {
		dx, dy := mag.X-x, mag.Y-y
		if d := math.Sqrt(dx*dx + dy*dy + m.Height*m.Height); d > 1e-10 {
			magPE -= mag.Strength / math.Pow(d, m.MagnetPower-1)
		}
	}
	return 0.5*(vx*vx+vy*vy) + pe + magPE
}

func (m *MagneticPendulum) ClosestMagnet(s dynamo.State) int {
	if len(s) < 2 {
		return -1
	}
	x, y, c, min := s[0], s[1], -1, math.MaxFloat64
	for i, mag := range m.Magnets {
		if d := (mag.X-x)*(mag.X-x) + (mag.Y-y)*(mag.Y-y); d < min {
			min, c = d, i
		}
	}
	return c
}

func (m *MagneticPendulum) GetParams() map[string]float64 {
	return map[string]float64{"height": m.Height, "damping": m.Damping, "gravity": m.Gravity, "magnetPower": m.MagnetPower}
}

func (m *MagneticPendulum) SetParam(n string, v float64) error {
	switch n {
	case "height":
		m.Height = v
	case "damping":
		m.Damping = v
	case "gravity":
		m.Gravity = v
	case "magnetPower":
		m.MagnetPower = v
	}
	return nil
}
