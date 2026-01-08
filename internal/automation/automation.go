package automation

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/san-kum/dynsim/internal/experiment"
	"github.com/san-kum/dynsim/internal/dynamo"
	"gopkg.in/yaml.v3"
)

// Scenario defines a scripted simulation sequence
type Scenario struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Steps       []ScenarioStep `yaml:"steps"`
}

// ScenarioStep is a single step in a scenario
type ScenarioStep struct {
	Model      string             `yaml:"model"`
	Integrator string             `yaml:"integrator"`
	Controller string             `yaml:"controller"`
	Duration   float64            `yaml:"duration"`
	Dt         float64            `yaml:"dt"`
	InitState  []float64          `yaml:"init_state"`
	Params     map[string]float64 `yaml:"params"`
	SaveAs     string             `yaml:"save_as"`
}

// LoadScenario loads a scenario from a YAML file
func LoadScenario(path string) (*Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var scenario Scenario
	if err := yaml.Unmarshal(data, &scenario); err != nil {
		return nil, err
	}

	return &scenario, nil
}

// RunScenario executes all steps in a scenario
func RunScenario(ctx context.Context, scenario *Scenario, registry *experiment.Registry) ([]dynamo.Result, error) {
	results := make([]dynamo.Result, 0, len(scenario.Steps))

	for i, step := range scenario.Steps {
		fmt.Printf("Running step %d/%d: %s\n", i+1, len(scenario.Steps), step.Model)

		dyn, err := registry.GetModel(step.Model)
		if err != nil {
			return results, fmt.Errorf("step %d: %w", i+1, err)
		}

		// Apply params if tunable
		if t, ok := dyn.(dynamo.Configurable); ok {
			for k, v := range step.Params {
				t.SetParam(k, v)
			}
		}

		integ, err := registry.GetIntegrator(step.Integrator)
		if err != nil {
			return results, fmt.Errorf("step %d: %w", i+1, err)
		}

		ctrl, err := registry.GetController(step.Controller, step.Params)
		if err != nil {
			return results, fmt.Errorf("step %d: %w", i+1, err)
		}

		cfg := experiment.Config{
			Model:      step.Model,
			Integrator: step.Integrator,
			Controller: step.Controller,
			InitState:  step.InitState,
			Dt:         step.Dt,
			Duration:   step.Duration,
		}

		exp := experiment.New(cfg)
		if err := exp.Setup(dyn, integ, ctrl, nil); err != nil {
			return results, fmt.Errorf("step %d setup: %w", i+1, err)
		}

		result, err := exp.Run(ctx)
		if err != nil {
			return results, fmt.Errorf("step %d run: %w", i+1, err)
		}

		results = append(results, *result)
	}

	return results, nil
}

// ParameterSweep runs simulations across a range of parameter values
type ParameterSweep struct {
	Model      string
	Integrator string
	ParamName  string
	ParamMin   float64
	ParamMax   float64
	NumSteps   int
	Duration   float64
	Dt         float64
	InitState  []float64
}

// SweepResult holds results from a parameter sweep
type SweepResult struct {
	ParamValue float64
	FinalState dynamo.State
	MaxEnergy  float64
	MinEnergy  float64
}

// RunSweep executes a parameter sweep
func RunSweep(ctx context.Context, sweep *ParameterSweep, registry *experiment.Registry) ([]SweepResult, error) {
	results := make([]SweepResult, 0, sweep.NumSteps)

	dyn, err := registry.GetModel(sweep.Model)
	if err != nil {
		return nil, err
	}

	integ, err := registry.GetIntegrator(sweep.Integrator)
	if err != nil {
		return nil, err
	}

	ctrl, err := registry.GetController("none", nil)
	if err != nil {
		return nil, err
	}

	tunable, ok := dyn.(dynamo.Configurable)
	if !ok {
		return nil, fmt.Errorf("model %s is not tunable", sweep.Model)
	}

	paramStep := (sweep.ParamMax - sweep.ParamMin) / float64(sweep.NumSteps-1)

	for i := 0; i < sweep.NumSteps; i++ {
		paramVal := sweep.ParamMin + float64(i)*paramStep
		tunable.SetParam(sweep.ParamName, paramVal)

		cfg := experiment.Config{
			Model:      sweep.Model,
			Integrator: sweep.Integrator,
			Controller: "none",
			InitState:  sweep.InitState,
			Dt:         sweep.Dt,
			Duration:   sweep.Duration,
		}

		exp := experiment.New(cfg)
		if err := exp.Setup(dyn, integ, ctrl, nil); err != nil {
			return nil, err
		}

		result, err := exp.Run(ctx)
		if err != nil {
			return nil, err
		}

		// Analyze result
		var maxE, minE float64
		var finalState dynamo.State
		if len(result.States) > 0 {
			finalState = result.States[len(result.States)-1]
			if ec, ok := dyn.(dynamo.Hamiltonian); ok {
				minE, maxE = ec.Energy(result.States[0]), ec.Energy(result.States[0])
				for _, s := range result.States {
					e := ec.Energy(s)
					if e > maxE {
						maxE = e
					}
					if e < minE {
						minE = e
					}
				}
			}
		}

		results = append(results, SweepResult{
			ParamValue: paramVal,
			FinalState: finalState,
			MaxEnergy:  maxE,
			MinEnergy:  minE,
		})

		fmt.Printf("Sweep %d/%d: %s=%.4f\n", i+1, sweep.NumSteps, sweep.ParamName, paramVal)
	}

	return results, nil
}

// MonteCarloConfig defines Monte Carlo simulation parameters
type MonteCarloConfig struct {
	Model        string
	Integrator   string
	BaseState    []float64
	Perturbation float64
	NumTrials    int
	Duration     float64
	Dt           float64
	Seed         int64
}

// MonteCarloResult holds statistics from Monte Carlo runs
type MonteCarloResult struct {
	TrialID    int
	InitState  dynamo.State
	FinalState dynamo.State
	Stable     bool // Did simulation remain bounded?
}

// RunMonteCarlo executes multiple trials with random perturbations
func RunMonteCarlo(ctx context.Context, cfg *MonteCarloConfig, registry *experiment.Registry) ([]MonteCarloResult, error) {
	results := make([]MonteCarloResult, 0, cfg.NumTrials)

	dyn, err := registry.GetModel(cfg.Model)
	if err != nil {
		return nil, err
	}

	integ, err := registry.GetIntegrator(cfg.Integrator)
	if err != nil {
		return nil, err
	}

	ctrl, err := registry.GetController("none", nil)
	if err != nil {
		return nil, err
	}

	rng := rand.New(rand.NewSource(cfg.Seed))
	if cfg.Seed == 0 {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	for trial := 0; trial < cfg.NumTrials; trial++ {
		// Perturb initial state
		initState := make([]float64, len(cfg.BaseState))
		for i, v := range cfg.BaseState {
			initState[i] = v + (rng.Float64()-0.5)*2*cfg.Perturbation
		}

		expCfg := experiment.Config{
			Model:      cfg.Model,
			Integrator: cfg.Integrator,
			Controller: "none",
			InitState:  initState,
			Dt:         cfg.Dt,
			Duration:   cfg.Duration,
		}

		exp := experiment.New(expCfg)
		if err := exp.Setup(dyn, integ, ctrl, nil); err != nil {
			return nil, err
		}

		result, err := exp.Run(ctx)
		if err != nil {
			return nil, err
		}

		// Check stability (final state bounded)
		stable := true
		var final dynamo.State
		if len(result.States) > 0 {
			final = result.States[len(result.States)-1]
			for _, v := range final {
				if v > 1e6 || v < -1e6 {
					stable = false
					break
				}
			}
		}

		results = append(results, MonteCarloResult{
			TrialID:    trial,
			InitState:  initState,
			FinalState: final,
			Stable:     stable,
		})

		if (trial+1)%10 == 0 {
			fmt.Printf("Monte Carlo: %d/%d trials complete\n", trial+1, cfg.NumTrials)
		}
	}

	return results, nil
}

// MonteCarloStats computes summary statistics from Monte Carlo results
func MonteCarloStats(results []MonteCarloResult) (stableCount int, unstableCount int) {
	for _, r := range results {
		if r.Stable {
			stableCount++
		} else {
			unstableCount++
		}
	}
	return
}
