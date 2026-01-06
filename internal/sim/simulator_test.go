package sim

import (
	"context"
	"math"
	"testing"
)

type testDynamics struct{}

func (t *testDynamics) Derivative(x State, u Control, time float64) State {
	return State{-x[0]}
}

func (t *testDynamics) StateDim() int   { return 1 }
func (t *testDynamics) ControlDim() int { return 0 }

type testIntegrator struct{}

func (t *testIntegrator) Step(dyn Dynamics, x State, u Control, time float64, dt float64) State {
	dx := dyn.Derivative(x, u, time)
	return State{x[0] + dt*dx[0]}
}

type testController struct{}

func (t *testController) Compute(x State, time float64) Control {
	return Control{}
}

func TestSimulatorRun(t *testing.T) {
	dyn := &testDynamics{}
	integ := &testIntegrator{}
	ctrl := &testController{}

	sim := New(dyn, integ, ctrl)

	cfg := Config{
		Dt:       0.1,
		Duration: 1.0,
	}

	x0 := State{1.0}
	result, err := sim.Run(context.Background(), x0, cfg)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if len(result.States) != 11 {
		t.Errorf("expected 11 states, got %d", len(result.States))
	}

	if len(result.Times) != 11 {
		t.Errorf("expected 11 times, got %d", len(result.Times))
	}

	finalState := result.States[len(result.States)-1][0]
	expected := 1.0 * math.Exp(-1.0)
	if math.Abs(finalState-expected) > 0.2 {
		t.Errorf("expected final state ~%.4f, got %.4f", expected, finalState)
	}
}

func TestSimulatorInvalidConfig(t *testing.T) {
	dyn := &testDynamics{}
	integ := &testIntegrator{}
	ctrl := &testController{}

	sim := New(dyn, integ, ctrl)

	tests := []struct {
		name string
		cfg  Config
	}{
		{"zero dt", Config{Dt: 0, Duration: 1.0}},
		{"negative dt", Config{Dt: -0.1, Duration: 1.0}},
		{"zero duration", Config{Dt: 0.1, Duration: 0}},
		{"negative duration", Config{Dt: 0.1, Duration: -1.0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x0 := State{1.0}
			_, err := sim.Run(context.Background(), x0, tt.cfg)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

type testMetric struct {
	count int
	sum   float64
}

func (t *testMetric) Name() string { return "test" }
func (t *testMetric) Observe(x State, u Control, time float64) {
	t.count++
	t.sum += x[0]
}
func (t *testMetric) Value() float64 {
	if t.count == 0 {
		return 0
	}
	return t.sum / float64(t.count)
}
func (t *testMetric) Reset() {
	t.count = 0
	t.sum = 0
}

func TestSimulatorMetrics(t *testing.T) {
	dyn := &testDynamics{}
	integ := &testIntegrator{}
	ctrl := &testController{}

	sim := New(dyn, integ, ctrl)

	metric := &testMetric{}
	sim.AddMetric(metric)

	cfg := Config{Dt: 0.1, Duration: 1.0}
	x0 := State{1.0}

	result, err := sim.Run(context.Background(), x0, cfg)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if _, ok := result.Metrics["test"]; !ok {
		t.Error("metric not found in result")
	}

	if metric.count != 10 {
		t.Errorf("expected 10 observations, got %d", metric.count)
	}
}
