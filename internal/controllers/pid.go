package controllers

import "github.com/san-kum/dynsim/internal/sim"

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

func (p *PID) Compute(x sim.State, t float64) sim.Control {
	if len(x) < 2 {
		return sim.Control{0}
	}

	err := p.Target - x[0]

	if p.first {
		p.prevErr = err
		p.prevT = t
		p.first = false
		return sim.Control{p.Kp * err}
	}

	dt := t - p.prevT
	if dt > 0 {
		p.integral += err * dt
		derivative := (err - p.prevErr) / dt

		u := p.Kp*err + p.Ki*p.integral + p.Kd*derivative

		p.prevErr = err
		p.prevT = t

		return sim.Control{u}
	}
	return sim.Control{p.Kp * err}
}
