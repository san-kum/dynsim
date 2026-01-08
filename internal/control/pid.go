package control

import "github.com/san-kum/dynsim/internal/dynamo"

type PID struct {
	Kp       float64
	Ki       float64
	Kd       float64
	Target   float64
	integral float64
	prevErr  float64
	prevT    float64
	first    bool
}

func NewPID(kp, ki, kd, target float64) *PID {
	return &PID{
		Kp:     kp,
		Ki:     ki,
		Kd:     kd,
		Target: target,
		first:  true,
	}
}

func (p *PID) Compute(x dynamo.State, t float64) dynamo.Control {
	if len(x) < 2 {
		return dynamo.Control{0}
	}

	err := p.Target - x[0]

	if p.first {
		p.prevErr = err
		p.prevT = t
		p.first = false
		return dynamo.Control{p.Kp * err}
	}

	dt := t - p.prevT
	if dt > 0 {
		p.integral += err * dt
		derivative := (err - p.prevErr) / dt

		u := p.Kp*err + p.Ki*p.integral + p.Kd*derivative

		p.prevErr = err
		p.prevT = t

		return dynamo.Control{u}
	}
	return dynamo.Control{p.Kp * err}
}

// Reset clears integral and derivative state
func (p *PID) Reset() {
	p.integral = 0
	p.prevErr = 0
	p.first = true
}

// GetParams returns tunable parameters for live adjustment
func (p *PID) GetParams() map[string]float64 {
	return map[string]float64{
		"Kp":     p.Kp,
		"Ki":     p.Ki,
		"Kd":     p.Kd,
		"Target": p.Target,
	}
}

// SetParam adjusts a PID parameter
func (p *PID) SetParam(name string, value float64) {
	switch name {
	case "Kp":
		p.Kp = value
	case "Ki":
		p.Ki = value
	case "Kd":
		p.Kd = value
	case "Target":
		p.Target = value
	}
}
