package export

import (
	"fmt"
	"strings"

	"github.com/san-kum/dynsim/internal/viz"
)

// CanvasToSVG converts a Braille canvas to SVG format
func CanvasToSVG(canvas *viz.Canvas, scale float64) string {
	if canvas == nil {
		return ""
	}

	width := float64(canvas.Width) * scale * 2   // 2 sub-pixels per char
	height := float64(canvas.Height) * scale * 4 // 4 sub-pixels per char

	var sb strings.Builder

	// SVG header
	sb.WriteString(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="%.0f" height="%.0f" viewBox="0 0 %.0f %.0f">
<rect width="100%%" height="100%%" fill="#0a0a0a"/>
<g fill="#00ff00">
`, width, height, width, height))

	// Braille dot-to-bit mapping
	pixelMap := [4][2]int{
		{0x01, 0x08},
		{0x02, 0x10},
		{0x04, 0x20},
		{0x40, 0x80},
	}

	dotRadius := scale * 0.4

	// Convert each braille character to dots
	for row := 0; row < canvas.Height; row++ {
		for col := 0; col < canvas.Width; col++ {
			r := canvas.Grid[row][col]
			if r < 0x2800 {
				continue
			}
			pattern := int(r - 0x2800)

			baseX := float64(col) * scale * 2
			baseY := float64(row) * scale * 4

			for dy := 0; dy < 4; dy++ {
				for dx := 0; dx < 2; dx++ {
					if pattern&pixelMap[dy][dx] != 0 {
						cx := baseX + float64(dx)*scale + scale/2
						cy := baseY + float64(dy)*scale + scale/2
						sb.WriteString(fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="%.1f"/>
`, cx, cy, dotRadius))
					}
				}
			}
		}
	}

	sb.WriteString("</g>\n</svg>")
	return sb.String()
}

// TrajectoryToSVG creates an SVG from trajectory data
func TrajectoryToSVG(points []struct{ X, Y float64 }, width, height int, strokeColor string) string {
	if len(points) < 2 {
		return ""
	}

	// Find bounds
	minX, maxX := points[0].X, points[0].X
	minY, maxY := points[0].Y, points[0].Y
	for _, p := range points {
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

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">
<rect width="100%%" height="100%%" fill="#0a0a0a"/>
<path fill="none" stroke="%s" stroke-width="1.5" d="M`,
		width, height, width, height, strokeColor))

	for i, p := range points {
		x := (p.X - minX) / rangeX * float64(width)
		y := float64(height) - (p.Y-minY)/rangeY*float64(height)

		if i == 0 {
			sb.WriteString(fmt.Sprintf("%.1f,%.1f", x, y))
		} else {
			sb.WriteString(fmt.Sprintf(" L%.1f,%.1f", x, y))
		}
	}

	sb.WriteString(`"/>
</svg>`)
	return sb.String()
}
