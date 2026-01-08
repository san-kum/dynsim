package physics

import "github.com/san-kum/dynsim/internal/dynamo"

// Wave implements a 1D wave equation using finite differences.
type Wave struct {
	N                              int
	Length, WaveSpeed, Damping, dx float64
}

func NewWave(n int) *Wave {
	if n < 3 {
		n = 3
	}
	return &Wave{n, 1.0, 1.0, 0.01, 1.0 / float64(n-1)}
}

func (w *Wave) StateDim() int   { return 2 * w.N }
func (w *Wave) ControlDim() int { return 0 }

func (w *Wave) Derive(s dynamo.State, _ dynamo.Control, _ float64) dynamo.State {
	n := w.N
	if len(s) < 2*n {
		return make(dynamo.State, 2*n)
	}
	dx, c2, h2 := dynamo.State(make([]float64, 2*n)), w.WaveSpeed*w.WaveSpeed, w.dx*w.dx
	for i := 0; i < n; i++ {
		dx[i] = s[n+i]
		if i == 0 || i == n-1 {
			dx[n+i] = -w.Damping * s[n+i]
		} else {
			dx[n+i] = c2*(s[i-1]-2*s[i]+s[i+1])/h2 - w.Damping*s[n+i]
		}
	}
	return dx
}

func (w *Wave) DefaultState() dynamo.State {
	s, c, amp := make(dynamo.State, 2*w.N), w.N/2, 0.5
	for i := 0; i < w.N; i++ {
		if i <= c {
			s[i] = amp * float64(i) / float64(c)
		} else {
			s[i] = amp * float64(w.N-1-i) / float64(w.N-1-c)
		}
	}
	return s
}

func (w *Wave) Energy(s dynamo.State) float64 {
	n, ke, pe, c2 := w.N, 0.0, 0.0, w.WaveSpeed*w.WaveSpeed
	if len(s) < 2*n {
		return 0
	}
	for i := 0; i < n; i++ {
		v := s[n+i]
		ke += 0.5 * v * v
		if i < n-1 {
			dudx := (s[i+1] - s[i]) / w.dx
			pe += 0.5 * c2 * dudx * dudx
		}
	}
	return ke + pe
}

func (w *Wave) GetParams() map[string]float64 {
	return map[string]float64{"waveSpeed": w.WaveSpeed, "damping": w.Damping, "length": w.Length}
}

func (w *Wave) SetParam(n string, v float64) error {
	switch n {
	case "waveSpeed":
		w.WaveSpeed = v
	case "damping":
		w.Damping = v
	case "length":
		w.Length, w.dx = v, v/float64(w.N-1)
	}
	return nil
}
