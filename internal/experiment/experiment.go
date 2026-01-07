package experiment

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/san-kum/dynsim/internal/sim"
)

type Config struct {
	Model      string
	Integrator string
	Controller string
	InitState  []float64
	Dt         float64
	Duration   float64
	Seed       int64
	Params     map[string]float64
}

type Experiment struct {
	cfg        Config
	simulator  *sim.Simulator
	randSource *rand.Rand
}

func New(cfg Config) *Experiment {
	return &Experiment{
		cfg:        cfg,
		randSource: rand.New(rand.NewSource(cfg.Seed)),
	}
}

func (e *Experiment) Setup(dyn sim.Dynamics, integrator sim.Integrator, controller sim.Controller, metrics []sim.Metric) error {
	e.simulator = sim.New(dyn, integrator, controller)
	for _, m := range metrics {
		e.simulator.AddMetric(m)
	}
	return nil
}

func (e *Experiment) Run(ctx context.Context) (*sim.Result, error) {
	if e.simulator == nil {
		return nil, fmt.Errorf("experiment not setup")
	}

	x0 := make(sim.State, len(e.cfg.InitState))
	copy(x0, e.cfg.InitState)

	simCfg := sim.Config{
		Dt:       e.cfg.Dt,
		Duration: e.cfg.Duration,
		Seed:     e.cfg.Seed,
	}

	return e.simulator.Run(ctx, x0, simCfg)
}

// GetSimulator returns the underlying simulator for adding observers
func (e *Experiment) GetSimulator() *sim.Simulator {
	return e.simulator
}
