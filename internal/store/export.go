package store

import (
	"encoding/json"
	"os"

	"github.com/san-kum/dynsim/internal/sim"
)

type ExportData struct {
	Model      string             `json:"model"`
	Integrator string             `json:"integrator"`
	Controller string             `json:"controller"`
	Dt         float64            `json:"dt"`
	Duration   float64            `json:"duration"`
	Steps      int                `json:"steps"`
	Times      []float64          `json:"times"`
	States     [][]float64        `json:"states"`
	Controls   [][]float64        `json:"controls"`
	Metrics    map[string]float64 `json:"metrics"`
}

func ExportJSON(path string, model, integrator, controller string, dt, duration float64, result *sim.Result) error {
	data := ExportData{
		Model:      model,
		Integrator: integrator,
		Controller: controller,
		Dt:         dt,
		Duration:   duration,
		Steps:      len(result.Times),
		Times:      result.Times,
		States:     make([][]float64, len(result.States)),
		Controls:   make([][]float64, len(result.Controls)),
		Metrics:    result.Metrics,
	}

	for i, s := range result.States {
		data.States[i] = s
	}
	for i, c := range result.Controls {
		data.Controls[i] = c
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func ExportJSONStdout(model, integrator, controller string, dt, duration float64, result *sim.Result) error {
	data := ExportData{
		Model:      model,
		Integrator: integrator,
		Controller: controller,
		Dt:         dt,
		Duration:   duration,
		Steps:      len(result.Times),
		Times:      result.Times,
		States:     make([][]float64, len(result.States)),
		Controls:   make([][]float64, len(result.Controls)),
		Metrics:    result.Metrics,
	}

	for i, s := range result.States {
		data.States[i] = s
	}
	for i, c := range result.Controls {
		data.Controls[i] = c
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
