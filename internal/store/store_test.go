package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/san-kum/dynsim/internal/sim"
)

func TestStoreSaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	st := New(tmpDir)

	if err := st.Init(); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	result := &sim.Result{
		States: []sim.State{
			{1.0, 0.0},
			{0.9, -0.1},
		},
		Controls: []sim.Control{
			{0.0},
		},
		Times: []float64{0.0, 0.01},
		Metrics: map[string]float64{
			"energy": 1.5,
		},
	}

	runID, err := st.Save("test", 0.01, 1.0, 42, "rk4", "none", result)
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	if runID == "" {
		t.Error("expected non-empty run id")
	}

	meta, err := st.Load(runID)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if meta.Model != "test" {
		t.Errorf("expected model 'test', got '%s'", meta.Model)
	}

	if meta.Seed != 42 {
		t.Errorf("expected seed 42, got %d", meta.Seed)
	}

	if meta.Metrics["energy"] != 1.5 {
		t.Errorf("expected energy 1.5, got %f", meta.Metrics["energy"])
	}

	states, times, err := st.LoadStates(runID)
	if err != nil {
		t.Fatalf("load states failed: %v", err)
	}

	if len(states) != 2 {
		t.Errorf("expected 2 states, got %d", len(states))
	}

	if len(times) != 2 {
		t.Errorf("expected 2 times, got %d", len(times))
	}
}

func TestStoreList(t *testing.T) {
	tmpDir := t.TempDir()
	st := New(tmpDir)

	if err := st.Init(); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	runs, err := st.List()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	if len(runs) != 0 {
		t.Errorf("expected 0 runs, got %d", len(runs))
	}

	result := &sim.Result{
		States:   []sim.State{{1.0}},
		Controls: []sim.Control{},
		Times:    []float64{0.0},
		Metrics:  map[string]float64{},
	}

	_, err = st.Save("test", 0.01, 1.0, 42, "rk4", "none", result)
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	runs, err = st.List()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	if len(runs) != 1 {
		t.Errorf("expected 1 run, got %d", len(runs))
	}
}

func TestStoreFileStructure(t *testing.T) {
	tmpDir := t.TempDir()
	st := New(tmpDir)

	if err := st.Init(); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	result := &sim.Result{
		States:   []sim.State{{1.0}},
		Controls: []sim.Control{},
		Times:    []float64{0.0},
		Metrics:  map[string]float64{},
	}

	runID, err := st.Save("test", 0.01, 1.0, 42, "rk4", "none", result)
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	runDir := filepath.Join(tmpDir, runID)
	metaPath := filepath.Join(runDir, "metadata.json")
	csvPath := filepath.Join(runDir, "states.csv")

	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Error("metadata.json not created")
	}

	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		t.Error("states.csv not created")
	}
}
