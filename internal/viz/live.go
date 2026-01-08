package viz

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/guptarohit/asciigraph"
	"github.com/san-kum/dynsim/internal/compute"
	"github.com/san-kum/dynsim/internal/dynamo"
)

const (
	width           = 80
	height          = 24
	historyCapacity = 600
)

// Snapshot stores state at a specific time for replay.
type Snapshot struct {
	State  dynamo.State
	Time   float64
	Energy float64
}

var (
	canvasStyle      = lipgloss.NewStyle().Padding(1, 2)
	statsStyle       = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(lipgloss.Color("240")).Padding(1, 2).Width(45)
	headerStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true).MarginBottom(1)
	labelStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Width(12)
	valueStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	activeParamStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	graphStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("49")).Padding(1, 0)
	helpStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).MarginTop(2)
)

type TickMsg time.Time

// Model contains simulation state, visualization buffers, and UI context.
type Model struct {
	dyn           dynamo.System
	integrator    dynamo.Integrator
	controller    dynamo.Controller
	state         dynamo.State
	u             dynamo.Control
	t, dt         float64
	width, height int
	canvas        *Canvas
	trail         []struct{ x, y int }
	trail3D       []Vec3
	camera3D      *Camera
	running       bool
	modelName     string
	stateHistory  [][]float64
	multiView     bool
	params        map[string]float64
	initialParams map[string]float64
	paramKeys     []string
	selected      int
	initialState  dynamo.State
	energyHistory []float64
	history       []Snapshot
	playHead      int
	recording     bool
	frames        []*image.Paletted
	showHelp      bool
}

// NewModel initializes the simulation and visualization state.
func NewModel(dyn dynamo.System, integ dynamo.Integrator, ctrl dynamo.Controller, initState []float64, dt float64, modelName string) Model {
	params := make(map[string]float64)
	if t, ok := dyn.(dynamo.Configurable); ok {
		for k, v := range t.GetParams() {
			params[k] = v
		}
	}
	keys := make([]string, 0, len(params))
	initialParams := make(map[string]float64)
	for k, v := range params {
		keys = append(keys, k)
		if v == 0 {
			v = 1e-6
		}
		initialParams[k] = v
	}
	sort.Strings(keys)

	return Model{
		dyn:           dyn,
		integrator:    integ,
		controller:    ctrl,
		state:         dynamo.State(initState),
		u:             make(dynamo.Control, dyn.ControlDim()),
		t:             0,
		dt:            dt,
		width:         width,
		height:        height,
		canvas:        NewCanvas(width, height),
		trail:         make([]struct{ x, y int }, 0, 100),
		trail3D:       make([]Vec3, 0, 500),
		camera3D:      NewCamera(),
		running:       true,
		modelName:     modelName,
		stateHistory:  make([][]float64, 0, historyCapacity),
		multiView:     false,
		params:        params,
		initialParams: initialParams,
		paramKeys:     keys,
		selected:      0,
		initialState:  cloneState(dynamo.State(initState)),
		energyHistory: make([]float64, 0, historyCapacity),
		history:       make([]Snapshot, 0, historyCapacity),
		playHead:      -1,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Tick(time.Second/60, func(t time.Time) tea.Msg { return TickMsg(t) })
}

// Update handles input events and steps the simulation.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case " ":
			m.running = !m.running
		case "r":
			m.reset()
		case "[":
			m.scrub(-1)
		case "]":
			m.scrub(1)
		case "tab":
			m.cycleParam()
		case "up", "k":
			m.adjustParam(1.05)
		case "down", "j":
			m.adjustParam(0.95)
		case "g":
			if m.recording {
				m.saveGIF()
				m.recording = false
				m.frames = nil
			} else {
				m.recording = true
				m.frames = make([]*image.Paletted, 0)
			}
		case "?":
			m.showHelp = !m.showHelp
		case "t":
			names := ThemeNames()
			for i, name := range names {
				if name == CurrentTheme.Name {
					SetTheme(names[(i+1)%len(names)])
					break
				}
			}
		case "x":
			m.camera3D.RotateX(0.1)
		case "X":
			m.camera3D.RotateX(-0.1)
		case "y":
			m.camera3D.RotateY(0.1)
		case "Y":
			m.camera3D.RotateY(-0.1)
		case "z":
			m.camera3D.RotateZ(0.1)
		case "Z":
			m.camera3D.RotateZ(-0.1)
		case "+", "=":
			m.camera3D.ZoomIn()
		case "-", "_":
			m.camera3D.ZoomOut()
		case "m":
			m.multiView = !m.multiView
		}
	case TickMsg:
		if m.running {
			if m.playHead == -1 {
				m.step()
			} else {
				m.playHead++
				if m.playHead >= len(m.history) {
					m.playHead = -1
				}
			}
		}
		m.draw()
		if m.recording {
			m.captureFrame()
		}
		return m, tea.Tick(time.Second/60, func(t time.Time) tea.Msg { return TickMsg(t) })
	}
	return m, nil
}

func (m *Model) cycleParam() {
	if len(m.paramKeys) == 0 {
		return
	}
	m.selected = (m.selected + 1) % len(m.paramKeys)
}

func (m *Model) adjustParam(factor float64) {
	if len(m.paramKeys) == 0 {
		return
	}
	key := m.paramKeys[m.selected]
	val := m.params[key]
	newVal := val * factor
	m.params[key] = newVal
	if t, ok := m.dyn.(dynamo.Configurable); ok {
		t.SetParam(key, newVal)
	}
}

// step advances the physics simulation.
func (m *Model) step() {
	m.u = m.controller.Compute(m.state, m.t)
	if adaptive, ok := m.integrator.(dynamo.AdaptiveIntegrator); ok {
		newState, suggestedDt, _ := adaptive.StepAdaptive(m.dyn, m.state, m.u, m.t, m.dt, 1e-6)
		m.state = newState
		m.t += m.dt
		if suggestedDt > 0.0001 && suggestedDt < 0.1 {
			m.dt = suggestedDt
		}
	} else {
		m.state = m.integrator.Step(m.dyn, m.state, m.u, m.t, m.dt)
		m.t += m.dt
	}

	energy := 0.0
	if e, ok := m.dyn.(dynamo.Hamiltonian); ok {
		energy = e.Energy(m.state)
	}
	m.energyHistory = append(m.energyHistory, energy)
	if len(m.energyHistory) > historyCapacity {
		m.energyHistory = m.energyHistory[1:]
	}

	stateCopy := make([]float64, len(m.state))
	copy(stateCopy, m.state)
	m.stateHistory = append(m.stateHistory, stateCopy)
	if len(m.stateHistory) > historyCapacity {
		m.stateHistory = m.stateHistory[1:]
	}

	snap := Snapshot{State: cloneState(m.state), Time: m.t, Energy: energy}
	m.history = append(m.history, snap)
	if len(m.history) > historyCapacity {
		m.history = m.history[1:]
	}
}

// scrub changes the playback position in history.
func (m *Model) scrub(dir int) {
	if m.playHead == -1 {
		if len(m.history) > 0 {
			m.playHead = len(m.history) - 1
			m.running = false
		} else {
			return
		}
	}
	m.playHead += dir
	if m.playHead < 0 {
		m.playHead = 0
	}
	if m.playHead >= len(m.history) {
		m.playHead = -1
	}
}

// reset restores the initial state and parameters.
func (m *Model) reset() {
	m.t = 0
	m.trail = m.trail[:0]
	m.state = cloneState(m.initialState)
	m.energyHistory = m.energyHistory[:0]
	m.history = m.history[:0]
	m.playHead = -1
	m.u = make([]float64, m.dyn.ControlDim())
	for k, v := range m.initialParams {
		m.params[k] = v
		if t, ok := m.dyn.(dynamo.Configurable); ok {
			t.SetParam(k, v)
		}
	}
}

func cloneState(s dynamo.State) dynamo.State {
	c := make(dynamo.State, len(s))
	copy(c, s)
	return c
}

// View renders the TUI interface.
func (m Model) View() string {
	state, t, energyHist, status := m.state, m.t, m.energyHistory, "RUNNING"
	if m.playHead >= 0 && m.playHead < len(m.history) {
		snap := m.history[m.playHead]
		state, t = snap.State, snap.Time
		status = fmt.Sprintf("REPLAY (%.1fs)", t-m.t)
		m.state, m.t = state, t
	}
	m.draw()
	canvasView := canvasStyle.Render(m.canvas.String())
	var s strings.Builder
	s.WriteString(headerStyle.Render(strings.ToUpper(m.modelName)) + "\n")
	if !m.running {
		if m.playHead != -1 {
			status = fmt.Sprintf("REPLAY PAUSED (%.1fs)", m.history[m.playHead].Time-m.history[len(m.history)-1].Time)
		} else {
			status = "PAUSED"
		}
	} else if m.playHead != -1 {
		status = fmt.Sprintf("REPLAYING (%.1fs)", m.history[m.playHead].Time-m.history[len(m.history)-1].Time)
	}
	s.WriteString(fmt.Sprintf("%s\n\n", status))
	if len(energyHist) > 1 {
		chart := asciigraph.Plot(energyHist, asciigraph.Height(4), asciigraph.Width(30), asciigraph.Caption("Energy"))
		s.WriteString(graphStyle.Render(chart) + "\n\n")
	}
	s.WriteString(labelStyle.Render("Time") + valueStyle.Render(fmt.Sprintf("%.2fs", t)) + "\n")
	energy := 0.0
	if len(energyHist) > 0 {
		energy = energyHist[len(energyHist)-1]
	}
	s.WriteString(labelStyle.Render("Energy") + valueStyle.Render(fmt.Sprintf("%.2f", energy)) + "\n")
	backend := compute.GetBackend()
	if backend.Available() {
		s.WriteString(labelStyle.Render("GPU") + valueStyle.Render("⚡ "+backend.Name()) + "\n")
	} else {
		s.WriteString(labelStyle.Render("GPU") + valueStyle.Render("CPU mode") + "\n")
	}
	s.WriteString("\nPARAMETERS\n")
	if len(m.params) > 0 {
		for i, k := range m.paramKeys {
			val, initial := m.params[k], m.initialParams[k]
			barWidth, ratio := 10, val/(2.0*initial)
			if ratio > 1 {
				ratio = 1
			} else if ratio < 0 {
				ratio = 0
			}
			filled := int(ratio * float64(barWidth))
			bar := "[" + strings.Repeat("=", filled) + strings.Repeat("-", barWidth-filled) + "]"
			line := fmt.Sprintf("%-10s %s %.2f", k, bar, val)
			if i == m.selected {
				s.WriteString(activeParamStyle.Render("> "+line) + "\n")
			} else {
				s.WriteString("  " + labelStyle.Render(line) + "\n")
			}
		}
	} else {
		s.WriteString(labelStyle.Render("  (none)") + "\n")
	}
	s.WriteString(helpStyle.Render("\n─────────────────────\nSP:Pause R:Reset Q:Quit\nT:Theme  G:Record ?:Help\n[ ]:Time-Travel ↑↓:Tune"))
	statsView := statsStyle.Render(s.String())
	mainView := lipgloss.JoinHorizontal(lipgloss.Top, canvasView, statsView)
	if m.showHelp {
		return `
╔══════════════════════════════════════╗
║           KEYBOARD SHORTCUTS          ║
╠══════════════════════════════════════╣
║  Space    - Pause/Resume simulation  ║
║  R        - Reset simulation         ║
║  Q        - Quit                     ║
║  Tab      - Cycle parameters         ║
║  Up/K     - Increase parameter (+5%) ║
║  Down/J   - Decrease parameter (-5%) ║
║  [        - Rewind (time travel)     ║
║  ]        - Forward (time travel)    ║
║  G        - Toggle GIF recording     ║
║  T        - Cycle themes             ║
║  ?        - Toggle this help         ║
╚══════════════════════════════════════╝
` + "\n\n" + mainView
	}
	return mainView
}

func (m *Model) clear() { m.canvas.Clear() }

// project maps model coordinates to screen space.
func (m *Model) project(x, y float64) (int, int, int, int) {
	cw, ch := m.width*2, m.height*4
	return cw, ch, cw / 2, ch / 2
}

// draw delegates rendering to specific model implementations.
func (m *Model) draw() {
	startState := m.state
	if m.playHead != -1 && m.playHead < len(m.history) {
		m.state = m.history[m.playHead].State
	}
	defer func() { m.state = startState }()
	m.clear()
	switch m.modelName {
	case "pendulum":
		m.drawPendulum()
	case "double_pendulum":
		m.drawDoublePendulum()
	case "cartpole":
		m.drawCartpole()
	case "spring_mass":
		m.drawSpring()
	case "drone":
		m.drawDrone()
	case "lorenz":
		m.drawLorenz()
	case "rossler":
		m.drawRossler()
	case "vanderpol":
		m.drawVanDerPol()
	case "threebody":
		m.drawThreeBody()
	case "coupled":
		m.drawCoupledPendulums()
	case "masschain":
		m.drawMassChain()
	case "gyroscope":
		m.drawGyroscope3D()
	case "wave":
		m.drawWave()
	case "doublewell":
		m.drawDoubleWell()
	case "duffing":
		m.drawDuffing()
	case "magnetic":
		m.drawMagneticPendulum()
	default:
		m.drawGeneric()
	}
}

func (m *Model) drawPendulum() {
	if len(m.state) < 2 {
		return
	}
	theta := m.state[0]
	_, ch, cx, _ := m.project(0, 0)
	cy, length := 8, float64(ch)*0.75
	bx, by := cx+int(length*math.Sin(theta)), cy+int(length*math.Cos(theta))
	m.trail = append(m.trail, struct{ x, y int }{bx, by})
	if len(m.trail) > 100 {
		m.trail = m.trail[1:]
	}
	for _, pt := range m.trail {
		m.canvas.Set(pt.x, pt.y)
	}
	m.canvas.Set(cx, cy)
	m.canvas.DrawLine(cx, cy, bx, by)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			m.canvas.Set(bx+dx, by+dy)
		}
	}
}

func (m *Model) drawDoublePendulum() {
	if len(m.state) < 4 {
		return
	}
	t1, t2 := m.state[0], m.state[1]
	_, ch, cx, _ := m.project(0, 0)
	cy, length := 8, float64(ch)*0.4
	b1x, b1y := cx+int(length*math.Sin(t1)), cy+int(length*math.Cos(t1))
	b2x, b2y := b1x+int(length*math.Sin(t2)), b1y+int(length*math.Cos(t2))
	m.trail = append(m.trail, struct{ x, y int }{b2x, b2y})
	if len(m.trail) > 200 {
		m.trail = m.trail[1:]
	}
	for _, pt := range m.trail {
		m.canvas.Set(pt.x, pt.y)
	}
	m.canvas.Set(cx, cy)
	m.canvas.DrawLine(cx, cy, b1x, b1y)
	m.canvas.Set(b1x, b1y)
	m.canvas.DrawLine(b1x, b1y, b2x, b2y)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			m.canvas.Set(b2x+dx, b2y+dy)
		}
	}
}

func (m *Model) drawCartpole() {
	if len(m.state) < 4 {
		return
	}
	pos, theta := m.state[0], m.state[2]
	_, ch, cx, _ := m.project(0, 0)
	groundY := ch - 12
	cartX := cx + int(pos*20)
	m.canvas.DrawLine(0, groundY+4, m.width*2, groundY+4)
	for dy := 0; dy < 4; dy++ {
		for dx := -6; dx <= 6; dx++ {
			m.canvas.Set(cartX+dx, groundY+dy)
		}
	}
	poleLen := float64(ch) * 0.6
	px, py := cartX+int(poleLen*math.Sin(theta)), groundY-int(poleLen*math.Cos(theta))
	m.canvas.DrawLine(cartX, groundY, px, py)
	m.canvas.Set(px, py)
}

func (m *Model) drawSpring() {
	if len(m.state) < 1 {
		return
	}
	pos := m.state[0]
	_, _, _, cy := m.project(0, 0)
	wallX := 20
	m.canvas.DrawLine(wallX, cy-10, wallX, cy+10)
	massX := wallX + 40 + int(pos*20)
	for dy := -4; dy <= 4; dy++ {
		for dx := -4; dx <= 4; dx++ {
			m.canvas.Set(massX+dx, cy+dy)
		}
	}
	numCoils, dist, prevX, prevY := 10, massX-wallX-4, wallX, cy
	step := float64(dist) / float64(numCoils)
	for i := 1; i <= numCoils; i++ {
		currX, amp := wallX+int(float64(i)*step), 6
		currY := cy
		if i%2 == 0 {
			currY -= amp
		} else {
			currY += amp
		}
		m.canvas.DrawLine(prevX, prevY, currX, currY)
		prevX, prevY = currX, currY
	}
	m.canvas.DrawLine(prevX, prevY, massX-4, cy)
}

func (m *Model) drawDrone() {
	if len(m.state) < 3 {
		return
	}
	px, py, theta := m.state[0], m.state[1], m.state[2]
	_, ch, cx, cy := m.project(0, 0)
	dx, dy := cx+int(px*8), cy-int(py*4)
	m.canvas.DrawLine(0, ch-4, m.width*2, ch-4)
	m.trail = append(m.trail, struct{ x, y int }{dx, dy})
	if len(m.trail) > 100 {
		m.trail = m.trail[1:]
	}
	for _, pt := range m.trail {
		m.canvas.Set(pt.x, pt.y)
	}
	arm, c, s := 12.0, math.Cos(theta), math.Sin(theta)
	lx, ly := dx-int(arm*c), dy-int(arm*s)
	rx, ry := dx+int(arm*c), dy+int(arm*s)
	m.canvas.DrawLine(lx, ly, rx, ry)
	m.canvas.DrawLine(lx-3, ly-2, lx+3, ly-2)
	m.canvas.DrawLine(rx-3, ry-2, rx+3, ry-2)
}

func (m *Model) drawGeneric() {
	_, _, _, cy := m.project(0, 0)
	barWidth, gap := 8, 4
	totalW := len(m.state) * (barWidth + gap)
	startX := (m.width*2 - totalW) / 2
	for i, v := range m.state {
		h, bx := int(v*10), startX+i*(barWidth+gap)
		if h > 0 {
			for y := cy; y > cy-h; y-- {
				for w := 0; w < barWidth; w++ {
					m.canvas.Set(bx+w, y)
				}
			}
		} else {
			for y := cy; y < cy-h; y++ {
				for w := 0; w < barWidth; w++ {
					m.canvas.Set(bx+w, y)
				}
			}
		}
	}
}

func (m *Model) drawLorenz() {
	if len(m.state) < 3 {
		return
	}
	x, y, z := m.state[0], m.state[1], m.state[2]
	scale := 0.04
	point := Vec3{x * scale, (z - 25) * scale, y * scale}
	m.trail3D = append(m.trail3D, point)
	if len(m.trail3D) > 500 {
		m.trail3D = m.trail3D[1:]
	}
	wf := NewWireframe()
	brightChars := []rune{'░', '▒', '▓', '█'}
	n := len(m.trail3D)
	for i := 1; i < n; i++ {
		age := float64(i) / float64(n)
		charIdx := int(age * float64(len(brightChars)-1))
		if charIdx >= len(brightChars) {
			charIdx = len(brightChars) - 1
		}
		wf.AddEdge(m.trail3D[i-1], m.trail3D[i], brightChars[charIdx])
	}
	wf.AddPoint(point, '●')
	m.camera3D.Position = Vec3{0, 0, 5}
	m.camera3D.Zoom = 1.0
	// slow rotate unless user is messing with it
	if m.camera3D.RotX == 0 && m.camera3D.RotZ == 0 {
		m.camera3D.RotY += 0.005
	}
	Render3D(m.canvas, wf, m.camera3D)
}

func (m *Model) drawRossler() {
	if len(m.state) < 3 {
		return
	}
	x, y, z := m.state[0], m.state[1], m.state[2]
	scale := 0.08
	point := Vec3{x * scale, z * scale * 0.5, y * scale}
	m.trail3D = append(m.trail3D, point)
	if len(m.trail3D) > 500 {
		m.trail3D = m.trail3D[1:]
	}
	wf := NewWireframe()
	for i := 1; i < len(m.trail3D); i++ {
		wf.AddEdge(m.trail3D[i-1], m.trail3D[i], '█')
	}
	wf.AddPoint(point, '●')
	m.camera3D.Position = Vec3{0, 0, 5}
	m.camera3D.Zoom = 1.0
	m.camera3D.RotY += 0.01
	Render3D(m.canvas, wf, m.camera3D)
}

func (m *Model) drawVanDerPol() {
	if len(m.state) < 2 {
		return
	}
	x, v := m.state[0], m.state[1]
	cw, ch, cx, cy := m.project(0, 0)
	scale := float64(ch) / 12.0
	px, py := cx+int(x*scale), cy-int(v*scale)
	if px < 0 {
		px = 0
	}
	if px >= cw {
		px = cw - 1
	}
	if py < 0 {
		py = 0
	}
	if py >= ch {
		py = ch - 1
	}
	m.trail = append(m.trail, struct{ x, y int }{px, py})
	if len(m.trail) > 500 {
		m.trail = m.trail[1:]
	}
	for _, pt := range m.trail {
		m.canvas.Set(pt.x, pt.y)
	}
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			m.canvas.Set(px+dx, py+dy)
		}
	}
}

func (m *Model) drawThreeBody() {
	if len(m.state) < 12 {
		return
	}
	cw, ch, cx, cy := m.project(0, 0)
	scale := float64(ch) / 6.0
	bodies := []struct{ x, y float64 }{{m.state[0], m.state[1]}, {m.state[4], m.state[5]}, {m.state[8], m.state[9]}}
	for _, body := range bodies {
		px, py := cx+int(body.x*scale), cy-int(body.y*scale)
		if px < 0 {
			px = 0
		}
		if px >= cw {
			px = cw - 1
		}
		if py < 0 {
			py = 0
		}
		if py >= ch {
			py = ch - 1
		}
		m.trail = append(m.trail, struct{ x, y int }{px, py})
		for dy := -2; dy <= 2; dy++ {
			for dx := -2; dx <= 2; dx++ {
				m.canvas.Set(px+dx, py+dy)
			}
		}
	}
	if len(m.trail) > 1000 {
		m.trail = m.trail[3:]
	}
	for _, pt := range m.trail {
		m.canvas.Set(pt.x, pt.y)
	}
}

func (m *Model) drawCoupledPendulums() {
	if len(m.state) < 4 {
		return
	}
	theta1, _, theta2, _ := m.state[0], m.state[1], m.state[2], m.state[3]
	_, ch, _, _ := m.project(0, 0)
	a1x, a2x, ay, l := 50, 110, 10, float64(ch)*0.5
	b1x, b1y := a1x+int(l*math.Sin(theta1)), ay+int(l*math.Cos(theta1))
	b2x, b2y := a2x+int(l*math.Sin(theta2)), ay+int(l*math.Cos(theta2))
	m.canvas.Set(a1x, ay)
	m.canvas.Set(a2x, ay)
	m.canvas.DrawLine(a1x, ay, b1x, b1y)
	m.canvas.DrawLine(a2x, ay, b2x, b2y)
	m.canvas.DrawLine(b1x, b1y, b2x, b2y)
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			m.canvas.Set(b1x+dx, b1y+dy)
			m.canvas.Set(b2x+dx, b2y+dy)
		}
	}
}

func (m *Model) drawMassChain() {
	n := len(m.state) / 2
	if n < 1 {
		return
	}
	cw, ch, _, cy := m.project(0, 0)
	spacing, scale := cw/(n+1), float64(ch)/4.0
	for i := 0; i < n; i++ {
		x, disp := (i+1)*spacing, m.state[i*2]
		y := cy + int(disp*scale)
		if y < 0 {
			y = 0
		}
		if y >= ch {
			y = ch - 1
		}
		for dy := -2; dy <= 2; dy++ {
			for dx := -2; dx <= 2; dx++ {
				m.canvas.Set(x+dx, y+dy)
			}
		}
		if i < n-1 {
			nextX, nextDisp := (i+2)*spacing, m.state[(i+1)*2]
			nextY := cy + int(nextDisp*scale)
			if nextY < 0 {
				nextY = 0
			}
			if nextY >= ch {
				nextY = ch - 1
			}
			m.canvas.DrawLine(x, y, nextX, nextY)
		}
	}
	m.canvas.DrawLine(0, 0, 0, ch-1)
	m.canvas.DrawLine(cw-1, 0, cw-1, ch-1)
}

func (m *Model) captureFrame() {
	charW, charH := 8, 16
	imgW, imgH := m.width*charW, m.height*charH
	img := image.NewPaletted(image.Rect(0, 0, imgW, imgH), color.Palette{color.Black, color.White})
	for y := 0; y < imgH; y++ {
		for x := 0; x < imgW; x++ {
			img.SetColorIndex(x, y, 0)
		}
	}
	for row := 0; row < m.height; row++ {
		for col := 0; col < m.width; col++ {
			r := m.canvas.Grid[row][col]
			if r < 0x2800 {
				continue
			}
			pattern := int(r - 0x2800)
			dotW, dotH := charW/2, charH/4
			baseX, baseY := col*charW, row*charH
			for dy := 0; dy < 4; dy++ {
				for dx := 0; dx < 2; dx++ {
					var bit int
					switch dy {
					case 0:
						bit = 1 << (dx * 3)
					case 1:
						bit = 2 << (dx * 3)
					case 2:
						bit = 4 << (dx * 3)
					case 3:
						if dx == 0 {
							bit = 0x40
						} else {
							bit = 0x80
						}
					}
					if pattern&bit != 0 {
						for py := 0; py < dotH; py++ {
							for px := 0; px < dotW; px++ {
								img.SetColorIndex(baseX+dx*dotW+px, baseY+dy*dotH+py, 1)
							}
						}
					}
				}
			}
		}
	}
	m.frames = append(m.frames, img)
}

func (m *Model) saveGIF() {
	if len(m.frames) == 0 {
		return
	}
	anim := gif.GIF{LoopCount: 0}
	for _, frame := range m.frames {
		anim.Image = append(anim.Image, frame)
		anim.Delay = append(anim.Delay, 2)
	}
	f, err := os.Create("simulation.gif")
	if err != nil {
		return
	}
	defer f.Close()
	gif.EncodeAll(f, &anim)
}

func (m *Model) drawGyroscope3D() {
	if len(m.state) < 6 {
		return
	}
	theta, phi, psi := m.state[3], m.state[4], m.state[5]
	wf := NewWireframe()
	topHeight, baseRadius := 1.0, 0.5
	numSegments := 12
	top := Vec3{0, topHeight, 0}
	for i := 0; i < numSegments; i++ {
		a1, a2 := float64(i)*2*math.Pi/float64(numSegments), float64(i+1)*2*math.Pi/float64(numSegments)
		p1, p2 := Vec3{baseRadius * math.Cos(a1), 0, baseRadius * math.Sin(a1)}, Vec3{baseRadius * math.Cos(a2), 0, baseRadius * math.Sin(a2)}
		p1, p2 = rotateVec3(p1, theta, phi, psi), rotateVec3(p2, theta, phi, psi)
		topRot := rotateVec3(top, theta, phi, psi)
		wf.AddEdge(p1, p2, '█')
		wf.AddEdge(p1, topRot, '█')
	}
	axisTop := rotateVec3(Vec3{0, topHeight + 0.3, 0}, theta, phi, psi)
	axisBot := rotateVec3(Vec3{0, -0.3, 0}, theta, phi, psi)
	wf.AddEdge(axisTop, axisBot, '│')
	camera := NewCamera()
	camera.Position = Vec3{0, 0, 5}
	camera.Zoom = 1.0
	Render3D(m.canvas, wf, camera)
}

func (m *Model) drawWave() {
	n := len(m.state) / 2
	if n < 2 {
		return
	}
	cw, ch, _, cy := m.project(0, 0)
	scaleX, scaleY := float64(cw-20)/float64(n), float64(ch)*0.4
	m.canvas.DrawLine(10, cy, cw-10, cy)
	prevX, prevY := 10, cy
	for i := 0; i < n; i++ {
		u := m.state[i]
		px, py := 10+int(float64(i)*scaleX), cy-int(u*scaleY)
		if py < 0 {
			py = 0
		}
		if py >= ch {
			py = ch - 1
		}
		if i > 0 {
			m.canvas.DrawLine(prevX, prevY, px, py)
		}
		prevX, prevY = px, py
	}
	for dy := -2; dy <= 2; dy++ {
		m.canvas.Set(10, cy+dy)
		m.canvas.Set(cw-10, cy+dy)
	}
}

func (m *Model) drawDoubleWell() {
	if len(m.state) < 2 {
		return
	}
	x, _ := m.state[0], m.state[1]
	cw, ch, cx, _ := m.project(0, 0)
	scaleX, scaleY := float64(cw)/6.0, float64(ch)*0.15
	groundY := ch - 20
	prevPx, prevPy := 0, 0
	for i := 0; i <= cw; i += 2 {
		xPos := (float64(i) - float64(cx)) / scaleX
		pot := math.Pow(xPos*xPos-1, 2)
		py := groundY - int(pot*scaleY)
		if py < 0 {
			py = 0
		}
		if i > 0 {
			m.canvas.DrawLine(prevPx, prevPy, i, py)
		}
		prevPx, prevPy = i, py
	}
	pX := cx + int(x*scaleX)
	pot := math.Pow(x*x-1, 2)
	pY := groundY - int(pot*scaleY) - 3
	m.trail = append(m.trail, struct{ x, y int }{pX, pY})
	if len(m.trail) > 50 {
		m.trail = m.trail[1:]
	}
	for _, pt := range m.trail {
		m.canvas.Set(pt.x, pt.y)
	}
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			m.canvas.Set(pX+dx, pY+dy)
		}
	}
}

func (m *Model) drawDuffing() {
	if len(m.state) < 2 {
		return
	}
	x, v := m.state[0], m.state[1]
	cw, ch, cx, cy := m.project(0, 0)
	scaleX, scaleY := float64(cw)/6.0, float64(ch)/6.0
	px, py := cx+int(x*scaleX), cy-int(v*scaleY)
	if px < 0 {
		px = 0
	}
	if px >= cw {
		px = cw - 1
	}
	if py < 0 {
		py = 0
	}
	if py >= ch {
		py = ch - 1
	}
	m.trail = append(m.trail, struct{ x, y int }{px, py})
	if len(m.trail) > 500 {
		m.trail = m.trail[1:]
	}
	for _, pt := range m.trail {
		m.canvas.Set(pt.x, pt.y)
	}
	m.canvas.DrawLine(0, cy, cw, cy)
	m.canvas.DrawLine(cx, 0, cx, ch)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			m.canvas.Set(px+dx, py+dy)
		}
	}
}

func (m *Model) drawMagneticPendulum() {
	if len(m.state) < 4 {
		return
	}
	x, y := m.state[0], m.state[1]
	cw, ch, cx, cy := m.project(0, 0)
	scale := float64(ch) / 5.0
	for i := 0; i < 3; i++ {
		a := float64(i) * 2 * math.Pi / 3
		mx, my := cx+int(1.5*math.Cos(a)*scale), cy-int(1.5*math.Sin(a)*scale)
		for dy := -3; dy <= 3; dy++ {
			for dx := -3; dx <= 3; dx++ {
				m.canvas.Set(mx+dx, my+dy)
			}
		}
	}
	px, py := cx+int(x*scale), cy-int(y*scale)
	if px < 0 {
		px = 0
	}
	if px >= cw {
		px = cw - 1
	}
	if py < 0 {
		py = 0
	}
	if py >= ch {
		py = ch - 1
	}
	m.trail = append(m.trail, struct{ x, y int }{px, py})
	if len(m.trail) > 300 {
		m.trail = m.trail[1:]
	}
	for _, pt := range m.trail {
		m.canvas.Set(pt.x, pt.y)
	}
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			m.canvas.Set(px+dx, py+dy)
		}
	}
}

func rotateVec3(v Vec3, th, ph, ps float64) Vec3 {
	ct, st := math.Cos(th), math.Sin(th)
	y1, z1 := v.Y*ct-v.Z*st, v.Y*st+v.Z*ct
	v.Y, v.Z = y1, z1
	cp, sp := math.Cos(ph), math.Sin(ph)
	x2, y2 := v.X*cp-v.Y*sp, v.X*sp+v.Y*cp
	v.X, v.Y = x2, y2
	cPs, sPs := math.Cos(ps), math.Sin(ps)
	x3, z3 := v.X*cPs+v.Z*sPs, -v.X*sPs+v.Z*cPs
	v.X, v.Z = x3, z3
	return v
}
