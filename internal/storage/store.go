package storage

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/san-kum/dynsim/internal/dynamo"
)

type Store struct {
	baseDir string
}

func New(baseDir string) *Store {
	return &Store{baseDir: baseDir}
}

func (s *Store) Init() error {
	return os.MkdirAll(s.baseDir, 0755)
}

type RunMetadata struct {
	ID         string             `json:"id"`
	Model      string             `json:"model"`
	Timestamp  time.Time          `json:"timestamp"`
	Seed       int64              `json:"seed"`
	Dt         float64            `json:"dt"`
	Duration   float64            `json:"duration"`
	Integrator string             `json:"integrator"`
	Controller string             `json:"controller"`
	Metrics    map[string]float64 `json:"metrics"`
}

func (s *Store) Save(model string, dt float64, duration float64, seed int64, integrator string, controller string, result *dynamo.Result) (string, error) {
	runID := fmt.Sprintf("%s_%d", model, time.Now().Unix())
	runDir := filepath.Join(s.baseDir, runID)

	if err := os.MkdirAll(runDir, 0755); err != nil {
		return "", err
	}

	meta := RunMetadata{
		ID:         runID,
		Model:      model,
		Timestamp:  time.Now(),
		Seed:       seed,
		Dt:         dt,
		Duration:   duration,
		Integrator: integrator,
		Controller: controller,
		Metrics:    result.Metrics,
	}

	metaPath := filepath.Join(runDir, "metadata.json")
	metaFile, err := os.Create(metaPath)
	if err != nil {
		return "", err
	}
	defer metaFile.Close()

	enc := json.NewEncoder(metaFile)
	enc.SetIndent("", "  ")
	if err := enc.Encode(meta); err != nil {
		return "", err
	}

	csvPath := filepath.Join(runDir, "states.csv")
	csvFile, err := os.Create(csvPath)
	if err != nil {
		return "", err
	}
	defer csvFile.Close()

	w := csv.NewWriter(csvFile)
	defer w.Flush()

	if len(result.States) == 0 {
		return runID, nil
	}

	header := []string{"time"}
	for i := range result.States[0] {
		header = append(header, fmt.Sprintf("x%d", i))
	}

	numControls := 0
	if len(result.Controls) > 0 && len(result.Controls[0]) > 0 {
		numControls = len(result.Controls[0])
		for i := 0; i < numControls; i++ {
			header = append(header, fmt.Sprintf("u%d", i))
		}
	}

	if err := w.Write(header); err != nil {
		return "", err
	}

	for i := range result.States {
		row := []string{strconv.FormatFloat(result.Times[i], 'f', 6, 64)}

		for _, val := range result.States[i] {
			row = append(row, strconv.FormatFloat(val, 'f', 6, 64))
		}

		if i < len(result.Controls) && len(result.Controls[i]) > 0 {
			for _, val := range result.Controls[i] {
				row = append(row, strconv.FormatFloat(val, 'f', 6, 64))
			}
		} else if numControls > 0 {
			for j := 0; j < numControls; j++ {
				row = append(row, "0")
			}
		}

		if err := w.Write(row); err != nil {
			return "", err
		}
	}

	return runID, nil
}

func (s *Store) List() ([]RunMetadata, error) {
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []RunMetadata{}, nil
		}
		return nil, err
	}

	runs := make([]RunMetadata, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		metaPath := filepath.Join(s.baseDir, entry.Name(), "metadata.json")
		data, err := os.ReadFile(metaPath)
		if err != nil {
			continue
		}

		var meta RunMetadata
		if err := json.Unmarshal(data, &meta); err != nil {
			continue
		}

		runs = append(runs, meta)
	}

	return runs, nil
}

func (s *Store) Load(runID string) (*RunMetadata, error) {
	metaPath := filepath.Join(s.baseDir, runID, "metadata.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}

	var meta RunMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

func (s *Store) LoadStates(runID string) ([][]float64, []float64, error) {
	csvPath := filepath.Join(s.baseDir, runID, "states.csv")
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	r := csv.NewReader(file)
	r.FieldsPerRecord = -1

	records, err := r.ReadAll()
	if err != nil {
		return nil, nil, err
	}

	if len(records) < 2 {
		return [][]float64{}, []float64{}, nil
	}

	times := make([]float64, 0, len(records)-1)
	states := make([][]float64, 0, len(records)-1)

	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) == 0 {
			continue
		}

		t, err := strconv.ParseFloat(record[0], 64)
		if err != nil {
			continue
		}
		times = append(times, t)

		state := make([]float64, 0)
		for j := 1; j < len(record); j++ {
			val, err := strconv.ParseFloat(record[j], 64)
			if err != nil {
				continue
			}
			state = append(state, val)
		}
		states = append(states, state)
	}

	return states, times, nil
}

