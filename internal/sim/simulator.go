package sim

import (
	"context"
	"fmt"
	"math"
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

func (s *Simulator) AddMetric(m Metric)     { s.metrics = append(s.metrics, m) }
func (s *Simulator) AddObserver(o Observer) { s.observers = append(s.observers, o) }

func (s *Simulator) Run(ctx context.Context, x0 State, cfg Config) (*Result, error) {
	if err := s.validateConfig(cfg); err != nil {
		return nil, err
	}

	steps := int(cfg.Duration / cfg.Dt)
	result := &Result{
		States:   make([]State, 0, steps+1),
		Controls: make([]Control, 0, steps),
		Times:    make([]float64, 0, steps+1),
		Metrics:  make(map[string]float64),
		Errors:   make([]error, 0),
	}

	for _, m := range s.metrics {
		m.Reset()
	}

	x := x0.Clone()
	t := 0.0
	dt := cfg.Dt

	result.States = append(result.States, x.Clone())
	result.Times = append(result.Times, t)

	initialEnergy := s.computeEnergy(x)

	for i := 0; i < steps; i++ {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		u := s.controller.Compute(x, t)

		for _, m := range s.metrics {
			m.Observe(x, u, t)
		}
		for _, obs := range s.observers {
			obs.OnStep(x, u, t)
		}

		var newX State
		var stepErr error

		if cfg.Adaptive {
			newX, dt, stepErr = s.adaptiveStep(x, u, t, dt, cfg)
		} else {
			newX = s.integrator.Step(s.dyn, x, u, t, dt)
		}

		if stepErr != nil {
			result.Errors = append(result.Errors, stepErr)
		}

		if cfg.ValidateState && !newX.IsValid() {
			err := SimError{Time: t, Step: i, Message: "invalid state (NaN/Inf)"}
			result.Errors = append(result.Errors, err)
			break
		}

		x = newX
		t += dt
		result.StepsTaken++

		result.States = append(result.States, x.Clone())
		result.Controls = append(result.Controls, u)
		result.Times = append(result.Times, t)
	}

	finalEnergy := s.computeEnergy(x)
	if initialEnergy != 0 {
		result.EnergyDrift = math.Abs(finalEnergy-initialEnergy) / math.Abs(initialEnergy)
	}

	for _, m := range s.metrics {
		result.Metrics[m.Name()] = m.Value()
	}

	return result, nil
}

func (s *Simulator) validateConfig(cfg Config) error {
	if cfg.Dt <= 0 {
		return fmt.Errorf("dt must be positive, got %f", cfg.Dt)
	}
	if cfg.Duration <= 0 {
		return fmt.Errorf("duration must be positive, got %f", cfg.Duration)
	}
	if cfg.Adaptive && cfg.Tolerance <= 0 {
		return fmt.Errorf("tolerance must be positive for adaptive stepping")
	}
	return nil
}

func (s *Simulator) computeEnergy(x State) float64 {
	if ec, ok := s.dyn.(EnergyComputer); ok {
		return ec.Energy(x)
	}
	return 0
}

func (s *Simulator) adaptiveStep(x State, u Control, t, dt float64, cfg Config) (State, float64, error) {
	if adaptive, ok := s.integrator.(AdaptiveIntegrator); ok {
		return adaptive.StepAdaptive(s.dyn, x, u, t, dt, cfg.Tolerance)
	}

	x1 := s.integrator.Step(s.dyn, x, u, t, dt)
	xHalf := s.integrator.Step(s.dyn, x, u, t, dt/2)
	x2 := s.integrator.Step(s.dyn, xHalf, u, t+dt/2, dt/2)

	err := x1.Sub(x2).Norm()

	if err > cfg.Tolerance && dt > cfg.MinDt {
		return s.adaptiveStep(x, u, t, dt/2, cfg)
	}

	if err < cfg.Tolerance/10 && dt < cfg.MaxDt {
		dt = math.Min(dt*2, cfg.MaxDt)
	}

	return x2, dt, nil
}

func (s *Simulator) RunWithCallback(ctx context.Context, x0 State, cfg Config, callback func(State, Control, float64) bool) error {
	if err := s.validateConfig(cfg); err != nil {
		return err
	}

	x := x0.Clone()
	t := 0.0
	dt := cfg.Dt

	for t < cfg.Duration {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		u := s.controller.Compute(x, t)

		if !callback(x, u, t) {
			return nil
		}

		x = s.integrator.Step(s.dyn, x, u, t, dt)
		t += dt

		if cfg.ValidateState && !x.IsValid() {
			return fmt.Errorf("invalid state at t=%.4f", t)
		}
	}

	return nil
}
