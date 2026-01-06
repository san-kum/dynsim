package experiment

import (
	"fmt"

	"github.com/san-kum/dynsim/internal/controllers"
	"github.com/san-kum/dynsim/internal/integrators"
	"github.com/san-kum/dynsim/internal/metrics"
	"github.com/san-kum/dynsim/internal/models"
	"github.com/san-kum/dynsim/internal/sim"
)

type Registry struct {
	models      map[string]func() sim.Dynamics
	integrators map[string]func() sim.Integrator
	controllers map[string]func(map[string]float64) sim.Controller
}

func NewRegistry() *Registry {
	r := &Registry{
		models:      make(map[string]func() sim.Dynamics),
		integrators: make(map[string]func() sim.Integrator),
		controllers: make(map[string]func(map[string]float64) sim.Controller),
	}

	r.models["pendulum"] = func() sim.Dynamics { return models.NewPendulum() }
	r.models["cartpole"] = func() sim.Dynamics { return models.NewCartPole() }
	r.models["nbody"] = func() sim.Dynamics { return models.NewNBody(3) }

	r.integrators["euler"] = func() sim.Integrator { return integrators.NewEuler() }
	r.integrators["rk4"] = func() sim.Integrator { return integrators.NewRK4() }
	r.integrators["verlet"] = func() sim.Integrator { return integrators.NewVerlet() }

	r.controllers["none"] = func(params map[string]float64) sim.Controller {
		dim := int(params["dim"])
		if dim == 0 {
			dim = 1
		}
		return controllers.NewNone(dim)
	}
	r.controllers["pid"] = func(params map[string]float64) sim.Controller {
		kp := params["kp"]
		ki := params["ki"]
		kd := params["kd"]
		target := params["target"]
		return controllers.NewPID(kp, ki, kd, target)
	}
	r.controllers["lqr"] = func(params map[string]float64) sim.Controller {
		return controllers.NewPendulumLQR()
	}

	return r
}

func (r *Registry) GetModel(name string) (sim.Dynamics, error) {
	fn, ok := r.models[name]
	if !ok {
		return nil, fmt.Errorf("unknown model: %s", name)
	}
	return fn(), nil
}

func (r *Registry) GetIntegrator(name string) (sim.Integrator, error) {
	fn, ok := r.integrators[name]
	if !ok {
		return nil, fmt.Errorf("unknown integrator: %s", name)
	}
	return fn(), nil
}

func (r *Registry) GetController(name string, params map[string]float64) (sim.Controller, error) {
	fn, ok := r.controllers[name]
	if !ok {
		return nil, fmt.Errorf("unknown controller: %s", name)
	}
	return fn(params), nil
}

func (r *Registry) ListModels() []string {
	names := make([]string, 0, len(r.models))
	for name := range r.models {
		names = append(names, name)
	}
	return names
}

func (r *Registry) DefaultMetrics(model string) []sim.Metric {
	metrics := []sim.Metric{
		metrics.NewEnergy(1.0, 1.0, 9.81),
		metrics.NewStability(10.0),
		metrics.NewControlEffort(),
	}
	return metrics
}
