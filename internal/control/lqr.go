package control

import "github.com/san-kum/dynsim/internal/dynamo"

type LQR struct {
	K      [][]float64
	Target dynamo.State
}

func NewLQR(k [][]float64, target dynamo.State) *LQR {
	return &LQR{K: k, Target: target}
}

func (l *LQR) Compute(x dynamo.State, t float64) dynamo.Control {
	u := make(dynamo.Control, len(l.K))
	for i := range u {
		for j := range x {
			target := 0.0
			if j < len(l.Target) {
				target = l.Target[j]
			}
			if j < len(l.K[i]) {
				u[i] -= l.K[i][j] * (x[j] - target)
			}
		}
	}
	return u
}

var (
	pendulumGains   = [][]float64{{31.62, 10.0}}
	cartpoleGains   = [][]float64{{-1.0, -1.73, 35.36, 8.94}}
	springGains     = [][]float64{{10.0, 6.32}}
	doublePendGains = [][]float64{{50.0, 40.0, 15.0, 10.0}}
)

func NewPendulumLQR() *LQR {
	return NewLQR(pendulumGains, dynamo.State{0, 0})
}

func NewCartPoleLQR() *LQR {
	return NewLQR(cartpoleGains, dynamo.State{0, 0, 0, 0})
}

func NewDroneLQR(targetY float64) *LQR {
	k := [][]float64{
		{0.0, -5.0, 10.0, 0.0, -3.5, 2.0},
		{0.0, -5.0, -10.0, 0.0, -3.5, -2.0},
	}
	return NewLQR(k, dynamo.State{0, targetY, 0, 0, 0, 0})
}

func NewDoublePendulumLQR() *LQR {
	return NewLQR(doublePendGains, dynamo.State{0, 0, 0, 0})
}

func NewSpringMassLQR() *LQR {
	return NewLQR(springGains, dynamo.State{0, 0})
}
