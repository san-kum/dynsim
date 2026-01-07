package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

const (
	DefaultDt       = 0.01
	DefaultDuration = 10.0
	DefaultTheta    = 0.5
	DefaultBodies   = 3
	DefaultY        = 5.0
	DefaultKp       = 10.0
	DefaultKi       = 0.1
	DefaultKd       = 5.0
)

type Config struct {
	Model            string           `yaml:"model"`
	Integrator       string           `yaml:"integrator"`
	Controller       string           `yaml:"controller"`
	Dt               float64          `yaml:"dt"`
	Duration         float64          `yaml:"duration"`
	Seed             int64            `yaml:"seed"`
	InitState        InitStateConfig  `yaml:"init_state"`
	ControllerParams ControllerConfig `yaml:"controller_params"`
}

type InitStateConfig struct {
	Theta     float64 `yaml:"theta"`
	Omega     float64 `yaml:"omega"`
	Theta2    float64 `yaml:"theta2"`
	Omega2    float64 `yaml:"omega2"`
	Pos       float64 `yaml:"pos"`
	Vel       float64 `yaml:"vel"`
	NumBodies int     `yaml:"num_bodies"`
	X         float64 `yaml:"x"`
	Y         float64 `yaml:"y"`
	VX        float64 `yaml:"vx"`
	VY        float64 `yaml:"vy"`
}

type ControllerConfig struct {
	Kp     float64 `yaml:"kp"`
	Ki     float64 `yaml:"ki"`
	Kd     float64 `yaml:"kd"`
	Target float64 `yaml:"target"`
}

func DefaultConfig() *Config {
	return &Config{
		Model:      "pendulum",
		Integrator: "rk4",
		Controller: "none",
		Dt:         DefaultDt,
		Duration:   DefaultDuration,
		InitState: InitStateConfig{
			Theta:     DefaultTheta,
			Theta2:    DefaultTheta,
			NumBodies: DefaultBodies,
			Y:         DefaultY,
		},
		ControllerParams: ControllerConfig{
			Kp: DefaultKp,
			Ki: DefaultKi,
			Kd: DefaultKd,
		},
	}
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (c *Config) GetInitState() []float64 {
	switch c.Model {
	case "cartpole":
		return []float64{c.InitState.Pos, c.InitState.Vel, c.InitState.Theta, c.InitState.Omega}
	case "double_pendulum":
		return []float64{c.InitState.Theta, c.InitState.Theta2, c.InitState.Omega, c.InitState.Omega2}
	case "spring_mass":
		return []float64{c.InitState.Pos, c.InitState.Vel}
	case "spring_chain":
		return []float64{c.InitState.Pos, 0, 0, c.InitState.Vel, 0, 0}
	case "drone":
		return []float64{c.InitState.X, c.InitState.Y, c.InitState.Theta, c.InitState.VX, c.InitState.VY, c.InitState.Omega}
	case "nbody":
		return nil
	default:
		return []float64{c.InitState.Theta, c.InitState.Omega}
	}
}

func (c *Config) GetControllerParams(controlDim int) map[string]float64 {
	return map[string]float64{
		"dim":    float64(controlDim),
		"kp":     c.ControllerParams.Kp,
		"ki":     c.ControllerParams.Ki,
		"kd":     c.ControllerParams.Kd,
		"target": c.ControllerParams.Target,
	}
}
