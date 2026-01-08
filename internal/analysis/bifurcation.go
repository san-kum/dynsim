package analysis

import (
	"github.com/san-kum/dynsim/internal/dynamo"
)

// BifurcationPoint represents a stable state for a given parameter value
type BifurcationPoint struct {
	Param  float64
	Values []float64 // Stable/periodic values found
}

// BifurcationDiagram sweeps a parameter and records stable states.
// This is useful for visualizing transitions to chaos.
//
// Parameters:
// - dyn: dynamics with Configurable interface
// - integ: integrator to use
// - paramName: name of parameter to sweep
// - paramMin, paramMax: range to sweep
// - paramSteps: number of parameter values to test
// - stateIndex: which state variable to record
// - dt, transient, record: timing parameters
func BifurcationDiagram(
	dyn dynamo.System,
	integ dynamo.Integrator,
	paramName string,
	paramMin, paramMax float64,
	paramSteps int,
	stateIndex int,
	x0 dynamo.State,
	dt, transient, record float64,
) []BifurcationPoint {
	tunable, ok := dyn.(dynamo.Configurable)
	if !ok {
		return nil
	}

	results := make([]BifurcationPoint, 0, paramSteps)
	if paramSteps <= 1 {
		paramSteps = 2 // Prevent division by zero
	}
	paramStep := (paramMax - paramMin) / float64(paramSteps-1)

	ctrl := make(dynamo.Control, dyn.ControlDim())

	for i := 0; i < paramSteps; i++ {
		param := paramMin + float64(i)*paramStep
		tunable.SetParam(paramName, param)

		// Reset state
		x := make(dynamo.State, len(x0))
		copy(x, x0)
		t := 0.0

		// Run transient (let system settle)
		for t < transient {
			x = integ.Step(dyn, x, ctrl, t, dt)
			t += dt
		}

		// Record stable values
		values := make([]float64, 0, 100)
		seen := make(map[int]bool)

		for t < transient+record {
			x = integ.Step(dyn, x, ctrl, t, dt)
			t += dt

			if stateIndex < len(x) {
				val := x[stateIndex]
				// Quantize to find distinct values
				key := int(val * 1000)
				if !seen[key] {
					seen[key] = true
					values = append(values, val)
				}
			}
		}

		results = append(results, BifurcationPoint{
			Param:  param,
			Values: values,
		})
	}

	// Restore original parameter
	if len(results) > 0 {
		tunable.SetParam(paramName, paramMin)
	}

	return results
}

// BifurcationToASCII converts bifurcation data to ASCII art
func BifurcationToASCII(data []BifurcationPoint, width, height int) string {
	if len(data) == 0 || width <= 0 || height <= 0 {
		return ""
	}

	// Find value range - need at least one valid value
	var minVal, maxVal float64
	foundFirst := false
	for _, p := range data {
		for _, v := range p.Values {
			if !foundFirst {
				minVal, maxVal = v, v
				foundFirst = true
			} else {
				if v < minVal {
					minVal = v
				}
				if v > maxVal {
					maxVal = v
				}
			}
		}
	}
	if !foundFirst {
		return "" // No values to plot
	}

	if maxVal == minVal {
		maxVal = minVal + 1
	}

	// Create canvas
	canvas := make([][]rune, height)
	for i := range canvas {
		canvas[i] = make([]rune, width)
		for j := range canvas[i] {
			canvas[i][j] = ' '
		}
	}

	// Plot points
	for i, p := range data {
		col := i * width / len(data)
		if col >= width {
			col = width - 1
		}

		for _, v := range p.Values {
			row := height - 1 - int((v-minVal)/(maxVal-minVal)*float64(height-1))
			if row >= 0 && row < height && col >= 0 && col < width {
				canvas[row][col] = '•'
			}
		}
	}

	// Convert to string
	result := ""
	for _, row := range canvas {
		result += string(row) + "\n"
	}
	return result
}
