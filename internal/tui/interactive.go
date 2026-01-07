package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/san-kum/dynsim/internal/controllers"
	"github.com/san-kum/dynsim/internal/integrators"
	"github.com/san-kum/dynsim/internal/models"
	"github.com/san-kum/dynsim/internal/sim"
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
	"pendulum":        "simple harmonic motion",
	"double_pendulum": "chaotic dynamics",
	"cartpole":        "balance control",
	"spring_mass":     "oscillator",
	"drone":           "2d quadrotor",
	"nbody":           "gravitational",
}

var stateLabels = map[string][]string{
	"pendulum":        {"θ", "ω"},
	"double_pendulum": {"θ₁", "θ₂", "ω₁", "ω₂"},
	"cartpole":        {"x", "ẋ", "θ", "ω"},
	"spring_mass":     {"x", "v"},
	"drone":           {"x", "y", "θ", "vx", "vy", "ω"},
	"nbody":           {"x₁", "y₁", "vx₁", "vy₁"},
}

type state int

const (
	stateMenu state = iota
	stateConfig
	stateSim
)

type model struct {
	state    state
	cursor   int
	models   []string
	selected string

	params      map[string]float64
	paramNames  []string
	paramCursor int
	editing     bool
	editBuf     string

	running   bool
	paused    bool
	simState  sim.State
	simTime   float64
	dynamics  sim.Dynamics
	dt        float64
	speed     float64
	trail     []trailPoint
	history   []float64
	lastFrame time.Time
	fps       float64

	width  int
	height int
}

type trailPoint struct {
	x, y     float64
	velocity float64
}

func NewInteractiveApp() *model {
	return &model{
		state:  stateMenu,
		models: []string{"pendulum", "double_pendulum", "cartpole", "spring_mass", "drone", "nbody"},
		params: map[string]float64{
			"theta": 0.5, "theta2": 0.5, "omega": 0.0, "omega2": 0.0,
			"pos": 0.0, "vel": 0.0, "dt": 0.01, "duration": 30.0,
		},
		paramNames: []string{"theta", "omega", "dt", "duration"},
		dt:         0.01,
		speed:      1.0,
		trail:      make([]trailPoint, 0, 100),
		history:    make([]float64, 0, 60),
		width:      80,
		height:     24,
	}
}

func (m model) Init() tea.Cmd { return nil }

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(16*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tickMsg:
		if m.state != stateSim {
			return m, nil
		}
		if m.running && !m.paused && m.dynamics != nil && m.simState != nil {
			now := time.Now()
			if !m.lastFrame.IsZero() {
				dt := now.Sub(m.lastFrame).Seconds()
				if dt > 0 {
					m.fps = 1.0 / dt
				}
			}
			m.lastFrame = now
			steps := int(m.speed)
			if steps < 1 {
				steps = 1
			}
			for i := 0; i < steps; i++ {
				m.step()
			}
		}
		if m.running && m.state == stateSim {
			return m, tick()
		}
		return m, nil
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
		return m.simKey(msg)
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
		m.state = stateConfig
		m.paramCursor = 0
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
			m.editing = false
			m.editBuf = ""
		case "escape":
			m.editing = false
			m.editBuf = ""
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
		m.editing = true
		m.editBuf = fmt.Sprintf("%.2f", m.params[m.paramNames[m.paramCursor]])
	case "s":
		m.start()
		m.state = stateSim
		return m, tea.Batch(tea.ClearScreen, tick())
	case "left", "h":
		m.params[m.paramNames[m.paramCursor]] -= 0.1
	case "right", "l":
		m.params[m.paramNames[m.paramCursor]] += 0.1
	}
	return m, nil
}

func (m model) simKey(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "q", "escape":
		m.running = false
		m.state = stateMenu
		m.reset()
		return m, tea.ClearScreen
	case " ", "p":
		m.paused = !m.paused
	case "r":
		m.start()
		return m, tea.ClearScreen
	case "c":
		m.running = false
		m.state = stateConfig
		m.reset()
		return m, tea.ClearScreen
	case "+", "=":
		m.speed = math.Min(m.speed*2, 16)
	case "-", "_":
		m.speed = math.Max(m.speed/2, 0.25)
	case "0":
		m.speed = 1.0
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
	}
	for _, name := range m.paramNames {
		if _, ok := m.params[name]; !ok {
			m.params[name] = 0.0
		}
	}
}

func (m *model) start() {
	if dt := m.params["dt"]; dt > 0 {
		m.dt = dt
	} else {
		m.dt = 0.01
	}
	m.trail = make([]trailPoint, 0, 100)
	m.history = make([]float64, 0, 60)
	m.simTime = 0
	m.speed = 1.0
	m.lastFrame = time.Time{}

	switch m.selected {
	case "pendulum":
		m.dynamics = models.NewPendulum()
		m.simState = sim.State{m.params["theta"], m.params["omega"]}
	case "double_pendulum":
		m.dynamics = models.NewDoublePendulum()
		m.simState = sim.State{m.params["theta"], m.params["theta2"], m.params["omega"], m.params["omega2"]}
	case "cartpole":
		m.dynamics = models.NewCartPole()
		m.simState = sim.State{m.params["pos"], m.params["vel"], m.params["theta"], m.params["omega"]}
	case "spring_mass":
		m.dynamics = models.NewSpringMass()
		m.simState = sim.State{m.params["pos"], m.params["vel"]}
	case "drone":
		m.dynamics = models.NewDrone()
		m.simState = sim.State{0, 5, m.params["theta"], 0, 0, m.params["omega"]}
	case "nbody":
		m.dynamics = models.NewNBody(3)
		m.simState = nbodyState(3)
	default:
		m.dynamics = models.NewPendulum()
		m.simState = sim.State{0.5, 0}
	}
	m.running = true
	m.paused = false
}

func (m *model) reset() {
	m.trail = nil
	m.history = nil
	m.dynamics = nil
	m.simState = nil
	m.simTime = 0
}

func (m *model) step() {
	if m.simTime >= m.params["duration"] {
		m.paused = true
		return
	}
	integrator := integrators.NewRK4()
	controller := controllers.NewNone(m.dynamics.ControlDim())
	u := controller.Compute(m.simState, m.simTime)
	m.simState = integrator.Step(m.dynamics, m.simState, u, m.simTime, m.dt)
	m.simTime += m.dt

	var tx, ty, vel float64
	switch m.selected {
	case "pendulum":
		tx, ty = m.simState[0], m.simState[1]
		vel = math.Abs(m.simState[1])
	case "double_pendulum":
		tx, ty = m.simState[0], m.simState[1]
		vel = math.Abs(m.simState[2]) + math.Abs(m.simState[3])
	case "drone":
		tx, ty = m.simState[0], m.simState[1]
		vel = math.Sqrt(m.simState[3]*m.simState[3] + m.simState[4]*m.simState[4])
	default:
		tx = m.simState[0]
		if len(m.simState) > 1 {
			ty = m.simState[1]
			vel = math.Abs(ty)
		}
	}
	m.trail = append(m.trail, trailPoint{tx, ty, vel})
	if len(m.trail) > 100 {
		m.trail = m.trail[1:]
	}
	if len(m.simState) > 0 {
		m.history = append(m.history, m.simState[0])
		if len(m.history) > 60 {
			m.history = m.history[1:]
		}
	}
}

func (m model) energy() (ke, pe float64) {
	if m.simState == nil || m.dynamics == nil {
		return 0, 0
	}
	switch m.selected {
	case "pendulum":
		if len(m.simState) >= 2 {
			theta, omega := m.simState[0], m.simState[1]
			ke = 0.5 * omega * omega
			pe = 9.81 * (1 - math.Cos(theta))
		}
	case "double_pendulum":
		if p, ok := m.dynamics.(*models.DoublePendulum); ok {
			total := p.Energy(m.simState)
			ke = math.Abs(total) * 0.5
			pe = math.Abs(total) * 0.5
		}
	case "spring_mass":
		if len(m.simState) >= 2 {
			x, v := m.simState[0], m.simState[1]
			ke = 0.5 * v * v
			pe = 5.0 * x * x
		}
	case "drone":
		if d, ok := m.dynamics.(*models.Drone); ok {
			total := d.Energy(m.simState)
			ke = math.Max(0, total*0.4)
			pe = math.Max(0, total*0.6)
		}
	}
	return
}

func (m model) View() string {
	switch m.state {
	case stateMenu:
		return m.viewMenu()
	case stateConfig:
		return m.viewConfig()
	case stateSim:
		return m.viewSim()
	}
	return ""
}

func (m model) viewMenu() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(dimmer.Render("    ╺━━━━━━━━━━━━━━━━━━━━━━━━╸") + "\n")
	b.WriteString("           " + cyan.Render("d y n s i m") + "\n")
	b.WriteString(dimmer.Render("    ╺━━━━━━━━━━━━━━━━━━━━━━━━╸") + "\n")
	b.WriteString("\n")

	for i, name := range m.models {
		desc := modelInfo[name]
		if i == m.cursor {
			b.WriteString("      " + cyan.Render("▸ ") + white.Render(fmt.Sprintf("%-16s", name)) + dim.Render(desc) + "\n")
		} else {
			b.WriteString("        " + dim.Render(fmt.Sprintf("%-16s", name)) + dimmer.Render(desc) + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(dim.Render("      ↑↓ select   enter start   q quit") + "\n")

	return b.String()
}

func (m model) viewConfig() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString("      " + cyan.Render(m.selected) + "  " + dim.Render(modelInfo[m.selected]) + "\n")
	b.WriteString(dimmer.Render("      "+strings.Repeat("─", 30)) + "\n\n")

	for i, name := range m.paramNames {
		val := fmt.Sprintf("%8.3f", m.params[name])
		if m.editing && i == m.paramCursor {
			val = fmt.Sprintf("%8s", m.editBuf+"▋")
		}
		if i == m.paramCursor {
			b.WriteString("      " + cyan.Render("▸ ") + white.Render(fmt.Sprintf("%-10s", name)) + magenta.Render(val) + "\n")
		} else {
			b.WriteString("        " + dim.Render(fmt.Sprintf("%-10s", name)) + dim.Render(val) + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(dim.Render("      ↑↓ select  ←→ adjust  enter edit  s start  esc back") + "\n")

	return b.String()
}

func (m model) viewSim() string {
	cw := m.width - 6
	ch := m.height - 12
	if cw < 50 {
		cw = 50
	}
	if ch < 12 {
		ch = 12
	}

	canvas := make([][]rune, ch)
	for i := range canvas {
		canvas[i] = make([]rune, cw)
		for j := range canvas[i] {
			canvas[i][j] = ' '
		}
	}

	switch m.selected {
	case "pendulum":
		m.drawPendulum(canvas, cw, ch)
	case "double_pendulum":
		m.drawDoublePendulum(canvas, cw, ch)
	case "cartpole":
		m.drawCartpole(canvas, cw, ch)
	case "spring_mass":
		m.drawSpring(canvas, cw, ch)
	case "drone":
		m.drawDrone(canvas, cw, ch)
	default:
		m.drawBars(canvas, cw, ch)
	}

	var b strings.Builder

	statusIcon := green.Render("●")
	statusText := green.Render("running")
	if m.paused {
		statusIcon = yellow.Render("○")
		statusText = yellow.Render("paused")
	}
	b.WriteString(fmt.Sprintf("\n   %s %s  %s\n",
		statusIcon, cyan.Render(m.selected), statusText))

	progress := m.simTime / m.params["duration"]
	if progress > 1 {
		progress = 1
	}
	barWidth := 36
	filled := int(progress * float64(barWidth))
	timeStr := fmt.Sprintf("%.1fs/%.0fs", m.simTime, m.params["duration"])
	bar := cyan.Render(strings.Repeat("━", filled)) + dimmer.Render(strings.Repeat("─", barWidth-filled))
	b.WriteString(fmt.Sprintf("   %s %s  %s\n\n", bar, dim.Render(timeStr), dim.Render(fmt.Sprintf("%.0ffps", m.fps))))

	for _, row := range canvas {
		b.WriteString("   " + string(row) + "\n")
	}

	ke, pe := m.energy()
	total := ke + pe
	if total > 0 {
		keRatio := ke / total
		energyWidth := 20
		keBar := int(keRatio * float64(energyWidth))
		peBar := energyWidth - keBar
		b.WriteString(fmt.Sprintf("\n   energy %s%s  %s %.1f  %s %.1f\n",
			green.Render(strings.Repeat("█", keBar)),
			yellow.Render(strings.Repeat("█", peBar)),
			green.Render("KE"), ke,
			yellow.Render("PE"), pe))
	}

	labels := stateLabels[m.selected]
	if len(labels) > 0 && len(m.simState) > 0 {
		var stateStr strings.Builder
		stateStr.WriteString("   ")
		for i, label := range labels {
			if i < len(m.simState) {
				stateStr.WriteString(dim.Render(label + "="))
				stateStr.WriteString(white.Render(fmt.Sprintf("%.2f", m.simState[i])))
				stateStr.WriteString("  ")
			}
			if i >= 3 {
				break
			}
		}
		b.WriteString(stateStr.String() + "\n")
	}

	if len(m.history) > 1 {
		spark := m.sparkline(m.history, 24)
		label := "θ"
		if len(stateLabels[m.selected]) > 0 {
			label = stateLabels[m.selected][0]
		}
		b.WriteString(fmt.Sprintf("   %s %s\n", dim.Render(label), cyan.Render(spark)))
	}

	b.WriteString("\n" + dim.Render("   space pause  ±speed  r reset  c config  q quit") + "\n")

	return b.String()
}

func (m model) sparkline(data []float64, width int) string {
	if len(data) == 0 {
		return ""
	}
	chars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	minVal, maxVal := data[0], data[0]
	for _, v := range data {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}
	rang := maxVal - minVal
	if rang == 0 {
		rang = 1
	}
	step := len(data) / width
	if step < 1 {
		step = 1
	}
	var sb strings.Builder
	for i := 0; i < width && i*step < len(data); i++ {
		v := data[i*step]
		idx := int((v - minVal) / rang * 7)
		if idx > 7 {
			idx = 7
		}
		if idx < 0 {
			idx = 0
		}
		sb.WriteRune(chars[idx])
	}
	return sb.String()
}

func (m model) trailChar(velocity, maxVel float64) rune {
	if maxVel == 0 {
		return '·'
	}
	ratio := velocity / maxVel
	if ratio < 0.25 {
		return '·'
	} else if ratio < 0.5 {
		return '∘'
	} else if ratio < 0.75 {
		return '○'
	}
	return '●'
}

func (m model) maxTrailVelocity() float64 {
	maxV := 0.0
	for _, pt := range m.trail {
		if pt.velocity > maxV {
			maxV = pt.velocity
		}
	}
	return maxV
}

func (m model) drawPendulum(canvas [][]rune, w, h int) {
	if len(m.simState) < 2 {
		return
	}
	theta := m.simState[0]
	px, py := w/2, 2
	length := float64(h) * 0.65
	bx := px + int(length*math.Sin(theta))
	by := py + int(length*math.Cos(theta))

	maxV := m.maxTrailVelocity()
	for _, pt := range m.trail {
		tx := px + int(length*math.Sin(pt.x))
		ty := py + int(length*math.Cos(pt.x))
		if tx >= 0 && tx < w && ty >= 0 && ty < h {
			canvas[ty][tx] = m.trailChar(pt.velocity, maxV)
		}
	}

	set(canvas, px, py, '▼', w, h)
	drawLine(canvas, w, h, px, py, bx, by, '│')
	set(canvas, bx, by, '⬤', w, h)
}

func (m model) drawDoublePendulum(canvas [][]rune, w, h int) {
	if len(m.simState) < 4 {
		return
	}
	t1, t2 := m.simState[0], m.simState[1]
	px, py := w/2, 1
	length := float64(h) * 0.38

	b1x := px + int(length*math.Sin(t1))
	b1y := py + int(length*math.Cos(t1))
	b2x := b1x + int(length*math.Sin(t2))
	b2y := b1y + int(length*math.Cos(t2))

	maxV := m.maxTrailVelocity()
	for _, pt := range m.trail {
		tx := px + int(length*math.Sin(pt.x)) + int(length*math.Sin(pt.y))
		ty := py + int(length*math.Cos(pt.x)) + int(length*math.Cos(pt.y))
		if tx >= 0 && tx < w && ty >= 0 && ty < h {
			canvas[ty][tx] = m.trailChar(pt.velocity, maxV)
		}
	}

	set(canvas, px, py, '▼', w, h)
	drawLine(canvas, w, h, px, py, b1x, b1y, '│')
	set(canvas, b1x, b1y, '●', w, h)
	drawLine(canvas, w, h, b1x, b1y, b2x, b2y, '│')
	set(canvas, b2x, b2y, '⬤', w, h)
}

func (m model) drawCartpole(canvas [][]rune, w, h int) {
	if len(m.simState) < 4 {
		return
	}
	pos, theta := m.simState[0], m.simState[2]
	gy := h - 3
	cx := w/2 + int(pos*10)
	if cx < 5 {
		cx = 5
	}
	if cx > w-5 {
		cx = w - 5
	}

	for x := 2; x < w-2; x++ {
		set(canvas, x, gy+1, '═', w, h)
	}
	for dx := -3; dx <= 3; dx++ {
		set(canvas, cx+dx, gy, '█', w, h)
	}
	set(canvas, cx-4, gy+1, '○', w, h)
	set(canvas, cx+4, gy+1, '○', w, h)

	plen := float64(h) * 0.55
	pex := cx + int(plen*math.Sin(theta))
	pey := gy - int(plen*math.Cos(theta))
	drawLine(canvas, w, h, cx, gy-1, pex, pey, '│')
	set(canvas, pex, pey, '⬤', w, h)
}

func (m model) drawSpring(canvas [][]rune, w, h int) {
	if len(m.simState) < 2 {
		return
	}
	pos := m.simState[0]
	cy := h / 2

	for y := cy - 3; y <= cy+3; y++ {
		set(canvas, 2, y, '█', w, h)
		set(canvas, 3, y, '█', w, h)
	}

	mx := w/4 + int(pos*12)
	if mx < 10 {
		mx = 10
	}
	if mx > w-8 {
		mx = w - 8
	}

	springLen := mx - 5
	coils := 8
	for i := 0; i < springLen; i++ {
		x := 5 + i
		phase := float64(i) / float64(springLen) * float64(coils) * 2 * math.Pi
		yOff := int(math.Sin(phase) * 1.5)
		set(canvas, x, cy+yOff, '~', w, h)
	}

	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			set(canvas, mx+dx, cy+dy, '█', w, h)
		}
	}
}

func (m model) drawDrone(canvas [][]rune, w, h int) {
	if len(m.simState) < 6 {
		return
	}
	x, y, theta := m.simState[0], m.simState[1], m.simState[2]

	for i := 1; i < w-1; i++ {
		set(canvas, i, h-1, '▀', w, h)
	}

	dx := w/2 + int(x*4)
	dy := h/2 - int(y*1.8)
	if dy < 1 {
		dy = 1
	}
	if dy > h-3 {
		dy = h - 3
	}

	maxV := m.maxTrailVelocity()
	for _, pt := range m.trail {
		tx := w/2 + int(pt.x*4)
		ty := h/2 - int(pt.y*1.8)
		if tx >= 0 && tx < w && ty >= 0 && ty < h-1 {
			canvas[ty][tx] = m.trailChar(pt.velocity, maxV)
		}
	}

	arm := 5.0
	lx := dx - int(arm*math.Cos(theta))
	ly := dy - int(arm*math.Sin(theta))
	rx := dx + int(arm*math.Cos(theta))
	ry := dy + int(arm*math.Sin(theta))

	drawLine(canvas, w, h, lx, ly, rx, ry, '─')
	set(canvas, dx, dy, '╋', w, h)
	set(canvas, lx, ly, '◉', w, h)
	set(canvas, rx, ry, '◉', w, h)
}

func (m model) drawBars(canvas [][]rune, w, h int) {
	cy := h / 2
	for x := 2; x < w-2; x++ {
		set(canvas, x, cy, '─', w, h)
	}
	if len(m.simState) == 0 {
		return
	}
	maxVal := 1.0
	for _, v := range m.simState {
		if math.Abs(v) > maxVal {
			maxVal = math.Abs(v)
		}
	}
	bw := (w - 8) / len(m.simState)
	if bw < 4 {
		bw = 4
	}
	for i, v := range m.simState {
		bx := 4 + i*bw
		bh := int((v / maxVal) * float64(h/3))
		if bh > 0 {
			for y := cy - 1; y >= cy-bh && y >= 1; y-- {
				set(canvas, bx, y, '█', w, h)
			}
		} else {
			for y := cy + 1; y <= cy-bh && y < h-1; y++ {
				set(canvas, bx, y, '█', w, h)
			}
		}
	}
}

func set(canvas [][]rune, x, y int, c rune, w, h int) {
	if x >= 0 && x < w && y >= 0 && y < h {
		canvas[y][x] = c
	}
}

func drawLine(canvas [][]rune, w, h, x1, y1, x2, y2 int, c rune) {
	dx := intAbs(x2 - x1)
	dy := intAbs(y2 - y1)
	sx, sy := 1, 1
	if x1 > x2 {
		sx = -1
	}
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy
	for {
		set(canvas, x1, y1, c, w, h)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func intAbs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func nbodyState(n int) sim.State {
	state := make(sim.State, n*4)
	for i := 0; i < n; i++ {
		angle := float64(i) * 2.0 * math.Pi / float64(n)
		state[i*4] = math.Cos(angle)
		state[i*4+1] = math.Sin(angle)
		state[i*4+2] = -math.Sin(angle) * 0.5
		state[i*4+3] = math.Cos(angle) * 0.5
	}
	return state
}

func RunInteractive() error {
	p := tea.NewProgram(NewInteractiveApp(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
