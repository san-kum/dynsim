package viz

import "github.com/charmbracelet/lipgloss"

// Theme defines color scheme for the TUI
type Theme struct {
	Name       string
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Accent     lipgloss.Color
	Background lipgloss.Color
	Text       lipgloss.Color
	Muted      lipgloss.Color
	Success    lipgloss.Color
	Warning    lipgloss.Color
	Error      lipgloss.Color
}

// Available themes
var (
	ThemeCyberpunk = Theme{
		Name:       "cyberpunk",
		Primary:    lipgloss.Color("#ff00ff"), // Magenta
		Secondary:  lipgloss.Color("#00ffff"), // Cyan
		Accent:     lipgloss.Color("#ffff00"), // Yellow
		Background: lipgloss.Color("#0a0a0a"),
		Text:       lipgloss.Color("#ffffff"),
		Muted:      lipgloss.Color("#666666"),
		Success:    lipgloss.Color("#00ff00"),
		Warning:    lipgloss.Color("#ff8800"),
		Error:      lipgloss.Color("#ff0000"),
	}

	ThemeRetroGreen = Theme{
		Name:       "retro",
		Primary:    lipgloss.Color("#00ff00"), // Green phosphor
		Secondary:  lipgloss.Color("#00cc00"),
		Accent:     lipgloss.Color("#88ff88"),
		Background: lipgloss.Color("#001100"),
		Text:       lipgloss.Color("#00ff00"),
		Muted:      lipgloss.Color("#005500"),
		Success:    lipgloss.Color("#88ff88"),
		Warning:    lipgloss.Color("#ffff00"),
		Error:      lipgloss.Color("#ff0000"),
	}

	ThemeMinimal = Theme{
		Name:       "minimal",
		Primary:    lipgloss.Color("#ffffff"),
		Secondary:  lipgloss.Color("#cccccc"),
		Accent:     lipgloss.Color("#0088ff"),
		Background: lipgloss.Color("#000000"),
		Text:       lipgloss.Color("#ffffff"),
		Muted:      lipgloss.Color("#888888"),
		Success:    lipgloss.Color("#00ff00"),
		Warning:    lipgloss.Color("#ffaa00"),
		Error:      lipgloss.Color("#ff0000"),
	}

	ThemeOcean = Theme{
		Name:       "ocean",
		Primary:    lipgloss.Color("#0077be"), // Ocean blue
		Secondary:  lipgloss.Color("#00a8cc"),
		Accent:     lipgloss.Color("#ffd700"),
		Background: lipgloss.Color("#001a33"),
		Text:       lipgloss.Color("#e0f0ff"),
		Muted:      lipgloss.Color("#4488aa"),
		Success:    lipgloss.Color("#00ff88"),
		Warning:    lipgloss.Color("#ffcc00"),
		Error:      lipgloss.Color("#ff4444"),
	}

	ThemeSunset = Theme{
		Name:       "sunset",
		Primary:    lipgloss.Color("#ff6b6b"), // Coral
		Secondary:  lipgloss.Color("#feca57"),
		Accent:     lipgloss.Color("#ff9ff3"),
		Background: lipgloss.Color("#2d1b2e"),
		Text:       lipgloss.Color("#fff5f5"),
		Muted:      lipgloss.Color("#8b6b8c"),
		Success:    lipgloss.Color("#5fd068"),
		Warning:    lipgloss.Color("#ffc048"),
		Error:      lipgloss.Color("#ff4757"),
	}

	// Default theme
	CurrentTheme = ThemeCyberpunk

	// All available themes
	Themes = []Theme{
		ThemeCyberpunk,
		ThemeRetroGreen,
		ThemeMinimal,
		ThemeOcean,
		ThemeSunset,
	}
)

// GetTheme returns a theme by name
func GetTheme(name string) Theme {
	for _, t := range Themes {
		if t.Name == name {
			return t
		}
	}
	return ThemeCyberpunk
}

// SetTheme changes the current theme
func SetTheme(name string) {
	CurrentTheme = GetTheme(name)
}

// ThemeNames returns list of available theme names
func ThemeNames() []string {
	names := make([]string, len(Themes))
	for i, t := range Themes {
		names[i] = t.Name
	}
	return names
}
