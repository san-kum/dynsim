package sim

import (
	"math"
	"testing"
)

func TestState_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		state State
		valid bool
	}{
		{"empty", State{}, true},
		{"normal", State{1.0, 2.0, 3.0}, true},
		{"zeros", State{0.0, 0.0}, true},
		{"with NaN", State{1.0, math.NaN()}, false},
		{"with +Inf", State{1.0, math.Inf(1)}, false},
		{"with -Inf", State{1.0, math.Inf(-1)}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.IsValid(); got != tt.valid {
				t.Errorf("IsValid() = %v, want %v", got, tt.valid)
			}
		})
	}
}

func TestState_Norm(t *testing.T) {
	tests := []struct {
		state    State
		expected float64
	}{
		{State{3, 4}, 5.0},
		{State{1, 0}, 1.0},
		{State{0, 0}, 0.0},
		{State{1, 1, 1, 1}, 2.0},
	}

	for _, tt := range tests {
		if got := tt.state.Norm(); math.Abs(got-tt.expected) > 1e-10 {
			t.Errorf("Norm(%v) = %v, want %v", tt.state, got, tt.expected)
		}
	}
}

func TestState_Arithmetic(t *testing.T) {
	a := State{1, 2, 3}
	b := State{4, 5, 6}

	sum := a.Add(b)
	if sum[0] != 5 || sum[1] != 7 || sum[2] != 9 {
		t.Errorf("Add failed: got %v", sum)
	}

	diff := b.Sub(a)
	if diff[0] != 3 || diff[1] != 3 || diff[2] != 3 {
		t.Errorf("Sub failed: got %v", diff)
	}

	scaled := a.Scale(2)
	if scaled[0] != 2 || scaled[1] != 4 || scaled[2] != 6 {
		t.Errorf("Scale failed: got %v", scaled)
	}
}

func TestStatePool(t *testing.T) {
	pool := NewStatePool(4)

	s1 := pool.Get()
	if len(s1) != 4 {
		t.Errorf("Pool returned wrong size: %d", len(s1))
	}

	s1[0] = 1.0
	s1[1] = 2.0
	pool.Put(s1)

	s2 := pool.Get()
	if s2[0] != 0 || s2[1] != 0 {
		t.Error("Pool did not reset state")
	}
}

func TestStatePool_GetAndCopy(t *testing.T) {
	pool := NewStatePool(3)
	src := State{1, 2, 3}

	copy := pool.GetAndCopy(src)
	if copy[0] != 1 || copy[1] != 2 || copy[2] != 3 {
		t.Errorf("GetAndCopy failed: got %v", copy)
	}

	copy[0] = 99
	if src[0] == 99 {
		t.Error("GetAndCopy did not create independent copy")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Dt <= 0 {
		t.Error("DefaultConfig has invalid Dt")
	}
	if cfg.Duration <= 0 {
		t.Error("DefaultConfig has invalid Duration")
	}
	if cfg.Tolerance <= 0 {
		t.Error("DefaultConfig has invalid Tolerance")
	}
}

func TestSimError(t *testing.T) {
	err := SimError{Time: 1.5, Step: 150, Message: "test error"}
	expected := "step 150 (t=1.5000): test error"
	if err.Error() != expected {
		t.Errorf("SimError.Error() = %q, want %q", err.Error(), expected)
	}
}
