package experiment

import (
	"fmt"

	"github.com/san-kum/dynsim/internal/control"
	"github.com/san-kum/dynsim/internal/integrators"
	"github.com/san-kum/dynsim/internal/metrics"
	"github.com/san-kum/dynsim/internal/physics"
	"github.com/san-kum/dynsim/internal/dynamo"
)

type ModelFactory func() dynamo.System
type IntegratorFactory func() dynamo.Integrator
type ControllerFactory func(map[string]float64) dynamo.Controller

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
	r.models["pendulum"] = func() dynamo.System { return physics.NewPendulum() }
	r.models["cartpole"] = func() dynamo.System { return physics.NewCartPole() }
	r.models["nbody"] = func() dynamo.System { return physics.NewNBody(3) }
	r.models["double_pendulum"] = func() dynamo.System { return physics.NewDoublePendulum() }
	r.models["spring_mass"] = func() dynamo.System { return physics.NewSpringMass() }
	r.models["spring_chain"] = func() dynamo.System { return physics.NewSpringMassChain(3) }
	r.models["drone"] = func() dynamo.System { return physics.NewDrone() }
}

func (r *Registry) registerIntegrators() {
	r.integrators["euler"] = func() dynamo.Integrator { return integrators.NewEuler() }
	r.integrators["rk4"] = func() dynamo.Integrator { return integrators.NewRK4() }
	r.integrators["rk45"] = func() dynamo.Integrator { return integrators.NewRK45() }
	r.integrators["verlet"] = func() dynamo.Integrator { return integrators.NewVerlet() }
	r.integrators["leapfrog"] = func() dynamo.Integrator { return integrators.NewLeapfrog() }
}

func (r *Registry) registerControllers() {
	r.controllers["none"] = func(p map[string]float64) dynamo.Controller {
		dim := int(p["dim"])
		if dim == 0 {
			dim = 1
		}
		return control.NewNone(dim)
	}

	r.controllers["pid"] = func(p map[string]float64) dynamo.Controller {
		return control.NewPID(p["kp"], p["ki"], p["kd"], p["target"])
	}

	r.controllers["lqr"] = func(p map[string]float64) dynamo.Controller {
		dim := int(p["dim"])
		switch dim {
		case 4:
			return control.NewCartPoleLQR()
		case 6:
			targetY := p["target"]
			if targetY == 0 {
				targetY = 5.0
			}
			return control.NewDroneLQR(targetY)
		default:
			return control.NewPendulumLQR()
		}
	}
}

func (r *Registry) GetModel(name string) (dynamo.System, error) {
	if fn, ok := r.models[name]; ok {
		return fn(), nil
	}
	return nil, fmt.Errorf("unknown model: %s", name)
}

func (r *Registry) GetIntegrator(name string) (dynamo.Integrator, error) {
	if fn, ok := r.integrators[name]; ok {
		return fn(), nil
	}
	return nil, fmt.Errorf("unknown integrator: %s", name)
}

func (r *Registry) GetController(name string, params map[string]float64) (dynamo.Controller, error) {
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

func (r *Registry) DefaultMetrics(model string) []dynamo.Metric {
	return []dynamo.Metric{
		metrics.NewEnergy(1.0, 1.0, 9.81),
		metrics.NewStability(10.0),
		metrics.NewControlEffort(),
	}
}
