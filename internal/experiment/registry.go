package experiment

import (
	"fmt"

	"github.com/san-kum/dynsim/internal/controllers"
	"github.com/san-kum/dynsim/internal/integrators"
	"github.com/san-kum/dynsim/internal/metrics"
	"github.com/san-kum/dynsim/internal/models"
	"github.com/san-kum/dynsim/internal/sim"
)

type ModelFactory func() sim.Dynamics
type IntegratorFactory func() sim.Integrator
type ControllerFactory func(map[string]float64) sim.Controller

type Registry struct {
	models      map[string]ModelFactory
	integrators map[string]IntegratorFactory
	controllers map[string]ControllerFactory
}

func NewRegistry() *Registry {
	r := &Registry{
		models:      make(map[string]ModelFactory),
		integrators: make(map[string]IntegratorFactory),
		controllers: make(map[string]ControllerFactory),
	}
	r.registerModels()
	r.registerIntegrators()
	r.registerControllers()
	return r
}

func (r *Registry) registerModels() {
	r.models["pendulum"] = func() sim.Dynamics { return models.NewPendulum() }
	r.models["cartpole"] = func() sim.Dynamics { return models.NewCartPole() }
	r.models["nbody"] = func() sim.Dynamics { return models.NewNBody(3) }
	r.models["double_pendulum"] = func() sim.Dynamics { return models.NewDoublePendulum() }
	r.models["spring_mass"] = func() sim.Dynamics { return models.NewSpringMass() }
	r.models["spring_chain"] = func() sim.Dynamics { return models.NewSpringMassChain(3) }
	r.models["drone"] = func() sim.Dynamics { return models.NewDrone() }
}

func (r *Registry) registerIntegrators() {
	r.integrators["euler"] = func() sim.Integrator { return integrators.NewEuler() }
	r.integrators["rk4"] = func() sim.Integrator { return integrators.NewRK4() }
	r.integrators["rk45"] = func() sim.Integrator { return integrators.NewRK45() }
	r.integrators["verlet"] = func() sim.Integrator { return integrators.NewVerlet() }
	r.integrators["leapfrog"] = func() sim.Integrator { return integrators.NewLeapfrog() }
}

func (r *Registry) registerControllers() {
	r.controllers["none"] = func(p map[string]float64) sim.Controller {
		dim := int(p["dim"])
		if dim == 0 {
			dim = 1
		}
		return controllers.NewNone(dim)
	}

	r.controllers["pid"] = func(p map[string]float64) sim.Controller {
		return controllers.NewPID(p["kp"], p["ki"], p["kd"], p["target"])
	}

	r.controllers["lqr"] = func(p map[string]float64) sim.Controller {
		dim := int(p["dim"])
		switch dim {
		case 4:
			return controllers.NewCartPoleLQR()
		case 6:
			targetY := p["target"]
			if targetY == 0 {
				targetY = 5.0
			}
			return controllers.NewDroneLQR(targetY)
		default:
			return controllers.NewPendulumLQR()
		}
	}
}

func (r *Registry) GetModel(name string) (sim.Dynamics, error) {
	if fn, ok := r.models[name]; ok {
		return fn(), nil
	}
	return nil, fmt.Errorf("unknown model: %s", name)
}

func (r *Registry) GetIntegrator(name string) (sim.Integrator, error) {
	if fn, ok := r.integrators[name]; ok {
		return fn(), nil
	}
	return nil, fmt.Errorf("unknown integrator: %s", name)
}

func (r *Registry) GetController(name string, params map[string]float64) (sim.Controller, error) {
	if fn, ok := r.controllers[name]; ok {
		return fn(params), nil
	}
	return nil, fmt.Errorf("unknown controller: %s", name)
}

func (r *Registry) ListModels() []string {
	names := make([]string, 0, len(r.models))
	for name := range r.models {
		names = append(names, name)
	}
	return names
}

func (r *Registry) DefaultMetrics(model string) []sim.Metric {
	return []sim.Metric{
		metrics.NewEnergy(1.0, 1.0, 9.81),
		metrics.NewStability(10.0),
		metrics.NewControlEffort(),
	}
}
