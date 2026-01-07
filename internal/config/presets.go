package config

var Presets = map[string]map[string]*Config{
	"pendulum": {
		"small": {
			Model: "pendulum", Integrator: "rk4", Dt: 0.01, Duration: 20.0,
			InitState: InitStateConfig{Theta: 0.2, Omega: 0.0},
		},
		"large": {
			Model: "pendulum", Integrator: "rk4", Dt: 0.01, Duration: 20.0,
			InitState: InitStateConfig{Theta: 2.5, Omega: 0.0},
		},
		"spinning": {
			Model: "pendulum", Integrator: "rk4", Dt: 0.01, Duration: 30.0,
			InitState: InitStateConfig{Theta: 0.1, Omega: 8.0},
		},
	},
	"double_pendulum": {
		"symmetric": {
			Model: "double_pendulum", Integrator: "rk4", Dt: 0.005, Duration: 30.0,
			InitState: InitStateConfig{Theta: 1.5, Theta2: 1.5, Omega: 0.0, Omega2: 0.0},
		},
		"chaos": {
			Model: "double_pendulum", Integrator: "rk4", Dt: 0.005, Duration: 60.0,
			InitState: InitStateConfig{Theta: 3.0, Theta2: 3.0, Omega: 0.0, Omega2: 0.0},
		},
		"gentle": {
			Model: "double_pendulum", Integrator: "rk4", Dt: 0.01, Duration: 30.0,
			InitState: InitStateConfig{Theta: 0.3, Theta2: 0.3, Omega: 0.0, Omega2: 0.0},
		},
	},
	"cartpole": {
		"balance": {
			Model: "cartpole", Integrator: "rk4", Controller: "lqr", Dt: 0.01, Duration: 30.0,
			InitState: InitStateConfig{Pos: 0.0, Vel: 0.0, Theta: 0.1, Omega: 0.0},
		},
		"recover": {
			Model: "cartpole", Integrator: "rk4", Controller: "lqr", Dt: 0.01, Duration: 30.0,
			InitState: InitStateConfig{Pos: 0.0, Vel: 0.0, Theta: 0.5, Omega: 0.0},
		},
		"freefall": {
			Model: "cartpole", Integrator: "rk4", Dt: 0.01, Duration: 10.0,
			InitState: InitStateConfig{Pos: 0.0, Vel: 0.0, Theta: 0.1, Omega: 0.0},
		},
	},
	"spring_mass": {
		"bounce": {
			Model: "spring_mass", Integrator: "rk4", Dt: 0.01, Duration: 20.0,
			InitState: InitStateConfig{Pos: 2.0, Vel: 0.0},
		},
		"fast": {
			Model: "spring_mass", Integrator: "rk4", Dt: 0.01, Duration: 10.0,
			InitState: InitStateConfig{Pos: 1.0, Vel: 5.0},
		},
	},
	"drone": {
		"hover": {
			Model: "drone", Integrator: "rk4", Controller: "lqr", Dt: 0.01, Duration: 30.0,
			InitState: InitStateConfig{X: 0, Y: 5, Theta: 0.0, VX: 0, VY: 0, Omega: 0},
		},
		"tilt": {
			Model: "drone", Integrator: "rk4", Dt: 0.01, Duration: 20.0,
			InitState: InitStateConfig{X: 0, Y: 5, Theta: 0.3, VX: 0, VY: 0, Omega: 0},
		},
		"drop": {
			Model: "drone", Integrator: "rk4", Dt: 0.01, Duration: 5.0,
			InitState: InitStateConfig{X: 0, Y: 10, Theta: 0.0, VX: 0, VY: 0, Omega: 0},
		},
	},
	"nbody": {
		"orbit": {
			Model: "nbody", Integrator: "leapfrog", Dt: 0.001, Duration: 50.0,
			InitState: InitStateConfig{NumBodies: 3},
		},
		"binary": {
			Model: "nbody", Integrator: "leapfrog", Dt: 0.001, Duration: 30.0,
			InitState: InitStateConfig{NumBodies: 2},
		},
	},
}

func GetPreset(model, preset string) *Config {
	modelPresets, ok := Presets[model]
	if !ok {
		return nil
	}
	cfg, ok := modelPresets[preset]
	if !ok {
		return nil
	}
	return cfg
}

func ListPresets(model string) []string {
	modelPresets, ok := Presets[model]
	if !ok {
		return nil
	}
	names := make([]string, 0, len(modelPresets))
	for name := range modelPresets {
		names = append(names, name)
	}
	return names
}
