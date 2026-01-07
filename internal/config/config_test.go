package config

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Model != "pendulum" {
		t.Errorf("expected model pendulum, got %s", cfg.Model)
	}
	if cfg.Dt <= 0 {
		t.Error("dt should be positive")
	}
	if cfg.Duration <= 0 {
		t.Error("duration should be positive")
	}
}

func TestGetPreset(t *testing.T) {
	cfg := GetPreset("pendulum", "small")
	if cfg == nil {
		t.Fatal("expected preset, got nil")
	}
	if cfg.InitState.Theta != 0.2 {
		t.Errorf("expected theta 0.2, got %f", cfg.InitState.Theta)
	}
}

func TestGetPreset_NotFound(t *testing.T) {
	cfg := GetPreset("pendulum", "nonexistent")
	if cfg != nil {
		t.Error("expected nil for nonexistent preset")
	}

	cfg = GetPreset("nonexistent", "small")
	if cfg != nil {
		t.Error("expected nil for nonexistent model")
	}
}

func TestListPresets(t *testing.T) {
	presets := ListPresets("pendulum")
	if len(presets) == 0 {
		t.Error("expected presets for pendulum")
	}

	presets = ListPresets("nonexistent")
	if presets != nil {
		t.Error("expected nil for nonexistent model")
	}
}

func TestGetInitState(t *testing.T) {
	tests := []struct {
		model    string
		expected int
	}{
		{"pendulum", 2},
		{"cartpole", 4},
		{"double_pendulum", 4},
		{"spring_mass", 2},
		{"drone", 6},
	}

	for _, tt := range tests {
		cfg := DefaultConfig()
		cfg.Model = tt.model
		state := cfg.GetInitState()
		if len(state) != tt.expected {
			t.Errorf("model %s: expected %d states, got %d", tt.model, tt.expected, len(state))
		}
	}
}
