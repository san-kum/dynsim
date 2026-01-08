package viz

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Premium style definitions for award-winning UI
var (
	// Glass panel effect with subtle border
	GlassPanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444466")).
			Padding(1, 2)

	// Gradient text simulation (alternating colors)
	GradientTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00ffff"))

	// Neon glow effect for selected items
	NeonGlow = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ff00ff")).
			Background(lipgloss.Color("#1a001a"))

	// Subtle muted text
	Subtle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666688"))

	// Status indicators
	StatusRunning = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00ff88"))

	StatusPaused = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffaa00"))

	StatusRecording = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ff4444")).
			Blink(true)

	// Metric value style
	MetricValue = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ccff")).
			Bold(true)

	// Metric label style
	MetricLabel = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888899"))

	// Key hint style
	KeyHint = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666688")).
		Italic(true)

	// Header with decorative line
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(lipgloss.Color("#444466"))

	// Sparkline bar colors
	SparkHigh = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff88"))
	SparkMid  = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffcc00"))
	SparkLow  = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff4444"))
)

// GradientText creates a gradient effect on text using color interpolation
func GradientText(text string, startColor, endColor lipgloss.Color) string {
	if len(text) == 0 {
		return ""
	}

	// Parse hex colors
	sr, sg, sb := parseHex(string(startColor))
	er, eg, eb := parseHex(string(endColor))

	var result strings.Builder
	n := len(text)

	for i, c := range text {
		t := float64(i) / float64(n-1)
		r := int(float64(sr) + t*float64(er-sr))
		g := int(float64(sg) + t*float64(eg-sg))
		b := int(float64(sb) + t*float64(eb-sb))

		color := lipgloss.Color(hexColor(r, g, b))
		style := lipgloss.NewStyle().Foreground(color)
		result.WriteString(style.Render(string(c)))
	}

	return result.String()
}

// AnimatedSpinner returns frame of animated spinner
func AnimatedSpinner(frame int) string {
	spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	return spinners[frame%len(spinners)]
}

// ProgressBar renders a beautiful progress bar
func ProgressBar(percent float64, width int) string {
	filled := int(percent * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	// Color gradient based on progress
	if percent > 0.8 {
		return SparkHigh.Render(bar)
	} else if percent > 0.4 {
		return SparkMid.Render(bar)
	}
	return SparkLow.Render(bar)
}

// SparklineChart renders a mini sparkline from values
func SparklineChart(values []float64, width int) string {
	if len(values) == 0 {
		return strings.Repeat("─", width)
	}

	// Sparkline characters from low to high
	chars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	// Find min/max
	min, max := values[0], values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	rng := max - min
	if rng == 0 {
		rng = 1
	}

	// Sample to fit width
	step := len(values) / width
	if step < 1 {
		step = 1
	}

	var result strings.Builder
	for i := 0; i < width && i*step < len(values); i++ {
		v := values[i*step]
		norm := (v - min) / rng
		idx := int(norm * float64(len(chars)-1))
		if idx >= len(chars) {
			idx = len(chars) - 1
		}
		if idx < 0 {
			idx = 0
		}

		// Color based on value
		c := chars[idx]
		if norm > 0.7 {
			result.WriteString(SparkHigh.Render(string(c)))
		} else if norm > 0.3 {
			result.WriteString(SparkMid.Render(string(c)))
		} else {
			result.WriteString(SparkLow.Render(string(c)))
		}
	}

	return result.String()
}

// BoxWithTitle renders a titled box
func BoxWithTitle(title, content string, width int) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00ffff"))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#444466")).
		Width(width).
		Padding(0, 1)

	header := "╭─ " + titleStyle.Render(title) + " " + strings.Repeat("─", width-len(title)-6) + "╮"
	return header + "\n" + box.Render(content)
}

// Decorative separator
func Separator(width int) string {
	mid := width / 2
	left := strings.Repeat("─", mid-3)
	right := strings.Repeat("─", width-mid-3)
	return Subtle.Render(left + " ◆ " + right)
}

// Helper functions
func parseHex(hex string) (r, g, b int) {
	if len(hex) != 7 || hex[0] != '#' {
		return 255, 255, 255
	}
	_, _ = parseHexByte(hex[1:3])
	r, _ = parseHexByte(hex[1:3])
	g, _ = parseHexByte(hex[3:5])
	b, _ = parseHexByte(hex[5:7])
	return
}

func parseHexByte(s string) (int, error) {
	var val int
	for _, c := range s {
		val *= 16
		if c >= '0' && c <= '9' {
			val += int(c - '0')
		} else if c >= 'a' && c <= 'f' {
			val += int(c - 'a' + 10)
		} else if c >= 'A' && c <= 'F' {
			val += int(c - 'A' + 10)
		}
	}
	return val, nil
}

func hexColor(r, g, b int) string {
	return "#" + hexByte(r) + hexByte(g) + hexByte(b)
}

func hexByte(v int) string {
	if v < 0 {
		v = 0
	}
	if v > 255 {
		v = 255
	}
	const hex = "0123456789abcdef"
	return string(hex[v/16]) + string(hex[v%16])
}
