package viz

import (
	"fmt"
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/san-kum/dynsim/internal/control"
	"github.com/san-kum/dynsim/internal/dynamo"
	"github.com/san-kum/dynsim/internal/integrators"
	"github.com/san-kum/dynsim/internal/physics"
)

var (
	cyan    = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	white   = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	dim     = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	dimmer  = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	green   = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	yellow  = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	magenta = lipgloss.NewStyle().Foreground(lipgloss.Color("213"))
)

var modelInfo = map[string]string{
	"pendulum": "simple harmonic motion", "double_pendulum": "chaotic dynamics", "cartpole": "balance control",
	"spring_mass": "oscillator", "drone": "2d quadrotor", "nbody": "gravitational", "lorenz": "butterfly attractor",
	"rossler": "spiral chaos", "vanderpol": "limit cycle oscillator", "threebody": "orbital chaos", "coupled": "energy transfer",
	"masschain": "wave propagation", "gyroscope": "rigid body rotation", "wave": "string vibration", "doublewell": "bistable potential",
	"duffing": "chaotic oscillator", "magnetic": "fractal basins",
}

const (
	stateMenu = iota
	stateConfig
	stateSim
)

type model struct {
	state, cursor int
	models        []string
	selected      string
	params        map[string]float64
	paramNames    []string
	paramCursor   int
	editing       bool
	editBuf       string
	running       bool
	paused        bool
	simState      dynamo.State
	simTime       float64
	dynamics      dynamo.System
	dt, speed     float64
	trail         []trailPoint
	history       []float64
	lastFrame     time.Time
	fps           float64
	width, height int
	liveModel     Model
}

type trailPoint struct {
	x, y, velocity float64
}

func NewInteractiveApp() *model {
	return &model{
		state:      stateMenu,
		models:     []string{"pendulum", "double_pendulum", "cartpole", "spring_mass", "drone", "nbody", "lorenz", "rossler", "vanderpol", "threebody", "coupled", "masschain", "gyroscope", "wave", "doublewell", "duffing", "magnetic"},
		params:     map[string]float64{"theta": 0.5, "theta2": 0.5, "omega": 0.0, "omega2": 0.0, "pos": 0.0, "vel": 0.0, "dt": 0.01, "duration": 30.0},
		paramNames: []string{"theta", "omega", "dt", "duration"},
		dt:         0.01, speed: 1.0, width: 80, height: 24,
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.width, m.height, m.liveModel.width = msg.Width, msg.Height, msg.Width
		return m, nil
	default:
		if m.state == stateSim {
			newLive, cmd := m.liveModel.Update(msg)
			m.liveModel = newLive.(Model)
			return m, cmd
		}
	}
	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (model, tea.Cmd) {
	switch m.state {
	case stateMenu:
		return m.menuKey(msg)
	case stateConfig:
		return m.configKey(msg)
	case stateSim:
		newLive, cmd := m.liveModel.Update(msg)
		m.liveModel = newLive.(Model)
		return m, cmd
	}
	return m, nil
}

func (m model) menuKey(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.models)-1 {
			m.cursor++
		}
	case "enter", " ":
		m.selected = m.models[m.cursor]
		m.state, m.paramCursor = stateConfig, 0
		m.setParamsForModel()
	}
	return m, nil
}

func (m model) configKey(msg tea.KeyMsg) (model, tea.Cmd) {
	if m.editing {
		switch msg.String() {
		case "enter":
			var val float64
			fmt.Sscanf(m.editBuf, "%f", &val)
			m.params[m.paramNames[m.paramCursor]] = val
			m.editing, m.editBuf = false, ""
		case "escape":
			m.editing, m.editBuf = false, ""
		case "backspace":
			if len(m.editBuf) > 0 {
				m.editBuf = m.editBuf[:len(m.editBuf)-1]
			}
		default:
			if len(msg.String()) == 1 {
				c := msg.String()[0]
				if (c >= '0' && c <= '9') || c == '.' || c == '-' {
					m.editBuf += string(c)
				}
			}
		}
		return m, nil
	}
	switch msg.String() {
	case "q", "escape":
		m.state = stateMenu
	case "up", "k":
		if m.paramCursor > 0 {
			m.paramCursor--
		}
	case "down", "j":
		if m.paramCursor < len(m.paramNames)-1 {
			m.paramCursor++
		}
	case "enter", " ":
		m.editing, m.editBuf = true, fmt.Sprintf("%.2f", m.params[m.paramNames[m.paramCursor]])
	case "s":
		cmd := m.start()
		return m, cmd
	case "left", "h":
		m.params[m.paramNames[m.paramCursor]] -= 0.1
	case "right", "l":
		m.params[m.paramNames[m.paramCursor]] += 0.1
	}
	return m, nil
}

func (m *model) setParamsForModel() {
	switch m.selected {
	case "pendulum":
		m.paramNames = []string{"theta", "omega", "dt", "duration"}
	case "double_pendulum":
		m.paramNames = []string{"theta", "theta2", "omega", "omega2", "dt", "duration"}
	case "cartpole":
		m.paramNames = []string{"pos", "vel", "theta", "omega", "dt", "duration"}
	case "spring_mass":
		m.paramNames = []string{"pos", "vel", "dt", "duration"}
	case "drone":
		m.paramNames = []string{"theta", "omega", "dt", "duration"}
	case "nbody":
		m.paramNames = []string{"dt", "duration"}
	case "lorenz", "rossler":
		m.paramNames = []string{"x", "y", "z", "dt", "duration"}
	case "vanderpol":
		m.paramNames = []string{"x", "v", "dt", "duration"}
	case "threebody", "masschain":
		m.paramNames = []string{"dt", "duration"}
	case "coupled":
		m.paramNames = []string{"theta1", "omega1", "theta2", "omega2", "dt", "duration"}
	}
	for _, name := range m.paramNames {
		if _, ok := m.params[name]; !ok {
			m.params[name] = 0.0
		}
	}
}

func (m *model) start() tea.Cmd {
	if dt := m.params["dt"]; dt > 0 {
		m.dt = dt
	} else {
		m.dt = 0.01
	}
	var dyn dynamo.System
	var state []float64
	switch m.selected {
	case "pendulum":
		dyn, state = physics.NewPendulum(), []float64{m.params["theta"], m.params["omega"]}
	case "double_pendulum":
		dyn, state = physics.NewDoublePendulum(), []float64{m.params["theta"], m.params["theta2"], m.params["omega"], m.params["omega2"]}
	case "cartpole":
		dyn, state = physics.NewCartPole(), []float64{m.params["pos"], m.params["vel"], m.params["theta"], m.params["omega"]}
	case "spring_mass":
		dyn, state = physics.NewSpringMass(), []float64{m.params["pos"], m.params["vel"]}
	case "drone":
		dyn, state = physics.NewDrone(), []float64{0, 5, m.params["theta"], 0, 0, m.params["omega"]}
	case "nbody":
		dyn, state = physics.NewNBody(3), nbodyState(3)
	case "lorenz":
		dyn, state = physics.NewLorenz(), []float64{m.params["x"], m.params["y"], m.params["z"]}
		if state[0] == 0 && state[1] == 0 && state[2] == 0 {
			state = []float64{1.0, 1.0, 1.0}
		}
	case "rossler":
		dyn, state = physics.NewRossler(), []float64{m.params["x"], m.params["y"], m.params["z"]}
		if state[0] == 0 && state[1] == 0 && state[2] == 0 {
			state = []float64{1.0, 1.0, 1.0}
		}
	case "vanderpol":
		dyn, state = physics.NewVanDerPol(), []float64{m.params["x"], m.params["v"]}
		if state[0] == 0 && state[1] == 0 {
			state = []float64{2.0, 0.0}
		}
	case "threebody":
		dyn = physics.NewThreeBody()
		state = dyn.(*physics.ThreeBody).DefaultState()
	case "coupled":
		dyn, state = physics.NewCoupledPendulums(), []float64{m.params["theta1"], m.params["omega1"], m.params["theta2"], m.params["omega2"]}
		if state[0] == 0 && state[2] == 0 {
			state = []float64{0.5, 0.0, 0.0, 0.0}
		}
	case "masschain":
		mc := physics.NewMassChain(20)
		dyn, state = mc, mc.DefaultState()
	case "gyroscope":
		g := physics.NewGyroscope()
		dyn, state = g, g.DefaultState()
	case "wave":
		w := physics.NewWave(30)
		dyn, state = w, w.DefaultState()
	case "doublewell":
		dw := physics.NewDoubleWell()
		dyn, state = dw, dw.DefaultState()
	case "duffing":
		df := physics.NewDuffing()
		dyn, state = df, df.DefaultState()
	case "magnetic":
		mp := physics.NewMagneticPendulum()
		dyn, state = mp, mp.DefaultState()
	default:
		dyn, state = physics.NewPendulum(), []float64{0.5, 0}
	}
	integ, ctrl := integrators.NewRK4(), control.NewNone(dyn.ControlDim())
	m.liveModel = NewModel(dyn, integ, ctrl, state, m.dt, m.selected)
	m.state = stateSim
	return m.liveModel.Init()
}

func (m model) View() string {
	switch m.state {
	case stateMenu:
		return m.viewMenu()
	case stateConfig:
		return m.viewConfig()
	case stateSim:
		return m.liveModel.View()
	}
	return ""
}

func (m model) viewMenu() string {
	var b strings.Builder
	h, sub := lipgloss.NewStyle().Foreground(lipgloss.Color("#00cccc")).Bold(true), lipgloss.NewStyle().Foreground(lipgloss.Color("#666688"))
	b.WriteString("\n\n    " + h.Render("DYNSIM") + "\n    " + sub.Render("physics simulation engine") + "\n    " + sub.Render("─────────────────────────") + "\n\n")
	for i, name := range m.models {
		desc := modelInfo[name]
		if len(desc) > 20 {
			desc = desc[:17] + "..."
		}
		if i == m.cursor {
			b.WriteString(fmt.Sprintf("    %s %s  %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("#00ffff")).Bold(true).Render("▸"), lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Bold(true).Render(fmt.Sprintf("%-16s", name)), lipgloss.NewStyle().Foreground(lipgloss.Color("#ff88ff")).Render(desc)))
		} else {
			b.WriteString(fmt.Sprintf("    %s  %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("#555566")).Render(fmt.Sprintf("  %-16s", name)), lipgloss.NewStyle().Foreground(lipgloss.Color("#444455")).Render(desc)))

		}
	}
	b.WriteString("\n    " + lipgloss.NewStyle().Foreground(lipgloss.Color("#00aaaa")).Bold(true).Render("j/k") + lipgloss.NewStyle().Foreground(lipgloss.Color("#555566")).Render(" navigate  ") + lipgloss.NewStyle().Foreground(lipgloss.Color("#00aaaa")).Bold(true).Render("enter") + lipgloss.NewStyle().Foreground(lipgloss.Color("#555566")).Render(" select  ") + lipgloss.NewStyle().Foreground(lipgloss.Color("#00aaaa")).Bold(true).Render("q") + lipgloss.NewStyle().Foreground(lipgloss.Color("#555566")).Render(" quit") + "\n")
	return b.String()
}

func (m model) viewConfig() string {
	var b strings.Builder
	h, sub := lipgloss.NewStyle().Foreground(lipgloss.Color("#00cccc")).Bold(true), lipgloss.NewStyle().Foreground(lipgloss.Color("#666688"))
	b.WriteString("\n\n    " + h.Render(strings.ToUpper(m.selected)) + "\n    " + sub.Render(modelInfo[m.selected]) + "\n    " + sub.Render("─────────────────────────") + "\n\n")
	for i, name := range m.paramNames {
		val := m.params[name]
		valStr := fmt.Sprintf("%8.3f", val)
		if m.editing && i == m.paramCursor {
			valStr = fmt.Sprintf("%8s", m.editBuf+"_")
		}
		if i == m.paramCursor {
			b.WriteString(fmt.Sprintf("    %s %s %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("#00ffff")).Bold(true).Render("▸"), lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Bold(true).Render(fmt.Sprintf("%-10s", name)), lipgloss.NewStyle().Foreground(lipgloss.Color("#ff88ff")).Bold(true).Render(valStr)))
		} else {
			b.WriteString(fmt.Sprintf("    %s %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("#555566")).Render(fmt.Sprintf("  %-10s", name)), lipgloss.NewStyle().Foreground(lipgloss.Color("#444455")).Render(valStr)))
		}
	}
	b.WriteString("\n    " + lipgloss.NewStyle().Foreground(lipgloss.Color("#00aaaa")).Bold(true).Render("j/k") + lipgloss.NewStyle().Foreground(lipgloss.Color("#555566")).Render(" select  ") + lipgloss.NewStyle().Foreground(lipgloss.Color("#00aaaa")).Bold(true).Render("h/l") + lipgloss.NewStyle().Foreground(lipgloss.Color("#555566")).Render(" adjust  ") + lipgloss.NewStyle().Foreground(lipgloss.Color("#00aaaa")).Bold(true).Render("s") + lipgloss.NewStyle().Foreground(lipgloss.Color("#555566")).Render(" start  ") + lipgloss.NewStyle().Foreground(lipgloss.Color("#00aaaa")).Bold(true).Render("esc") + lipgloss.NewStyle().Foreground(lipgloss.Color("#555566")).Render(" back") + "\n")
	return b.String()
}

func nbodyState(n int) dynamo.State {
	state := make(dynamo.State, n*4)
	for i := 0; i < n; i++ {
		angle := float64(i) * 2.0 * math.Pi / float64(n)
		state[i*4], state[i*4+1], state[i*4+2], state[i*4+3] = math.Cos(angle), math.Sin(angle), -math.Sin(angle)*0.5, math.Cos(angle)*0.5
	}
	return state
}

func RunInteractive() error { return tea.NewProgram(NewInteractiveApp(), tea.WithAltScreen()).Start() }
