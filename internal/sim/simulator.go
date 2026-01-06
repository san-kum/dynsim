package sim

import (
	"context"
	"fmt"
)

type Simulator struct {
	dyn        Dynamics
	integrator Integrator
	controller Controller
	metrics    []Metric
	observers  []Observer
}

func New(dyn Dynamics, integrator Integrator, controller Controller) *Simulator {
	return &Simulator{
		dyn:        dyn,
		integrator: integrator,
		controller: controller,
		metrics:    make([]Metric, 0),
		observers:  make([]Observer, 0),
	}
}

func (s *Simulator) AddMetric(m Metric) {
	s.metrics = append(s.metrics, m)
}

func (s *Simulator) AddObserver(o Observer) {
	s.observers = append(s.observers, o)
}

func (s *Simulator) Run(ctx context.Context, x0 State, cfg Config) (*Result, error) {
	if cfg.Dt <= 0 {
		return nil, fmt.Errorf("dt must be positive, got %f", cfg.Dt)
	}

	if cfg.Duration <= 0 {
		return nil, fmt.Errorf("duration must be positive, got %f", cfg.Duration)
	}

	steps := int(cfg.Duration / cfg.Dt)
	states := make([]State, 0, steps+1)
	controls := make([]Control, 0, steps)
	times := make([]float64, 0, steps+1)

	for _, m := range s.metrics {
		m.Reset()
	}

	x := x0.Clone()
	t := 0.0

	states = append(states, x.Clone())
	times = append(times, t)

	for i := 0; i < steps; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		u := s.controller.Compute(x, t)

		for _, m := range s.metrics {
			m.Observe(x, u, t)
		}
		for _, obs := range s.observers {
			obs.OnStep(x, u, t)
		}

		x = s.integrator.Step(s.dyn, x, u, t, cfg.Dt)
		t += cfg.Dt

		states = append(states, x.Clone())
		controls = append(controls, u)
		times = append(times, t)
	}

	metricValues := make(map[string]float64)
	for _, m := range s.metrics {
		metricValues[m.Name()] = m.Value()
	}

	return &Result{
		States:   states,
		Controls: controls,
		Times:    times,
		Metrics:  metricValues,
	}, nil
}
