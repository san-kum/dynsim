package analysis

import (
	"math"
	"strings"

	"github.com/san-kum/dynsim/internal/dynamo"
)

// PhasePortrait2D holds data for a 2D phase space plot
type PhasePortrait2D struct {
	XIndex, YIndex int
	Points         []struct{ X, Y float64 }
}

// GeneratePhasePortrait runs a simulation and records phase space trajectory
func GeneratePhasePortrait(
	dyn dynamo.System,
	integ dynamo.Integrator,
	x0 dynamo.State,
	xIdx, yIdx int,
	dt, duration float64,
) *PhasePortrait2D {
	if xIdx >= len(x0) || yIdx >= len(x0) {
		return nil
	}

	portrait := &PhasePortrait2D{
		XIndex: xIdx,
		YIndex: yIdx,
		Points: make([]struct{ X, Y float64 }, 0, int(duration/dt)),
	}

	x := make(dynamo.State, len(x0))
	copy(x, x0)
	ctrl := make(dynamo.Control, dyn.ControlDim())
	t := 0.0

	for t < duration {
		x = integ.Step(dyn, x, ctrl, t, dt)
		t += dt

		portrait.Points = append(portrait.Points, struct{ X, Y float64 }{
			X: x[xIdx],
			Y: x[yIdx],
		})
	}

	return portrait
}

// PhasePortraitToASCII converts phase portrait to ASCII art
func PhasePortraitToASCII(portrait *PhasePortrait2D, width, height int) string {
	if portrait == nil || len(portrait.Points) == 0 {
		return ""
	}

	// Find bounds
	minX, maxX := portrait.Points[0].X, portrait.Points[0].X
	minY, maxY := portrait.Points[0].Y, portrait.Points[0].Y

	for _, p := range portrait.Points {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}

	// Add padding
	rangeX := maxX - minX
	rangeY := maxY - minY
	if rangeX == 0 {
		rangeX = 1
	}
	if rangeY == 0 {
		rangeY = 1
	}
	minX -= rangeX * 0.1
	maxX += rangeX * 0.1
	minY -= rangeY * 0.1
	maxY += rangeY * 0.1
	rangeX = maxX - minX
	rangeY = maxY - minY

	// Create canvas
	canvas := make([][]rune, height)
	for i := range canvas {
		canvas[i] = make([]rune, width)
		for j := range canvas[i] {
			canvas[i][j] = ' '
		}
	}

	// Plot points
	for _, p := range portrait.Points {
		col := int((p.X - minX) / rangeX * float64(width-1))
		row := height - 1 - int((p.Y-minY)/rangeY*float64(height-1))

		if row >= 0 && row < height && col >= 0 && col < width {
			canvas[row][col] = '•'
		}
	}

	// Draw axes if they cross the visible area
	if minX <= 0 && maxX >= 0 {
		col := int((0 - minX) / rangeX * float64(width-1))
		for row := 0; row < height; row++ {
			if col >= 0 && col < width && canvas[row][col] == ' ' {
				canvas[row][col] = '│'
			}
		}
	}
	if minY <= 0 && maxY >= 0 {
		row := height - 1 - int((0-minY)/rangeY*float64(height-1))
		for col := 0; col < width; col++ {
			if row >= 0 && row < height && canvas[row][col] == ' ' {
				canvas[row][col] = '─'
			}
		}
	}

	// Convert to string
	var sb strings.Builder
	for _, row := range canvas {
		sb.WriteString(string(row))
		sb.WriteRune('\n')
	}
	return sb.String()
}

// PoincareSection records points when a trajectory crosses a plane
type PoincareSection struct {
	Points []struct{ X, Y float64 }
}

// GeneratePoincareSection creates a Poincaré section by recording state
// when a specified variable crosses a threshold
func GeneratePoincareSection(
	dyn dynamo.System,
	integ dynamo.Integrator,
	x0 dynamo.State,
	crossIdx int, // Index of variable that triggers recording
	threshold float64, // Value to cross
	recordX, recordY int, // Which state variables to record
	dt, duration float64,
) *PoincareSection {
	if crossIdx >= len(x0) || recordX >= len(x0) || recordY >= len(x0) {
		return nil
	}

	section := &PoincareSection{
		Points: make([]struct{ X, Y float64 }, 0),
	}

	x := make(dynamo.State, len(x0))
	copy(x, x0)
	ctrl := make(dynamo.Control, dyn.ControlDim())
	t := 0.0
	prevVal := x[crossIdx]

	for t < duration {
		x = integ.Step(dyn, x, ctrl, t, dt)
		t += dt
		currVal := x[crossIdx]

		// Detect positive-going crossing
		if prevVal < threshold && currVal >= threshold {
			// Interpolate for better accuracy
			frac := (threshold - prevVal) / (currVal - prevVal)
			if math.IsNaN(frac) || math.IsInf(frac, 0) {
				frac = 0.5
			}

			section.Points = append(section.Points, struct{ X, Y float64 }{
				X: x[recordX],
				Y: x[recordY],
			})
		}

		prevVal = currVal
	}

	return section
}

// PoincareSectionToASCII converts section data to ASCII plot
func PoincareSectionToASCII(section *PoincareSection, width, height int) string {
	if section == nil || len(section.Points) == 0 {
		return "No crossings detected"
	}

	// Use same logic as phase portrait
	portrait := &PhasePortrait2D{Points: section.Points}
	return PhasePortraitToASCII(portrait, width, height)
}
