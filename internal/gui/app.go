package gui

import (
	"fmt"
	"os"
	"sort"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/san-kum/dynsim/internal/control"
	"github.com/san-kum/dynsim/internal/dynamo"
	"github.com/san-kum/dynsim/internal/integrators"
	"github.com/san-kum/dynsim/internal/physics"
)

// Theme Colors (Monochrome Hyper-Minimalist)
var (
	ColBg      = rl.NewColor(10, 10, 10, 255)    // Deep Black
	ColAccent  = rl.NewColor(180, 180, 180, 255) // Soft White
	ColSelect  = rl.NewColor(255, 255, 255, 255) // Bright White
	ColText    = rl.NewColor(140, 140, 140, 255) // Neutral Gray
	ColTextDim = rl.NewColor(60, 60, 60, 255)    // Dark Gray (Subtle)
	ColGrid    = rl.NewColor(30, 30, 30, 255)    // Barely visible grid
)

type App struct {
	Dyn       dynamo.System
	Integ     dynamo.Integrator
	Ctrl      dynamo.Controller
	State     dynamo.State
	ModelName string
	Time      float64
	Dt        float64
	Camera    rl.Camera3D
	Running   bool
	InMenu    bool
	InConfig  bool
	Models    []string
	Selected  int
	Params    map[string]float64
	ParamKeys []string
	ParamSel  int
	Font      rl.Font
}

func initWindow() {
	rl.InitWindow(1280, 720, "dynsim")
	rl.SetTargetFPS(60)
	rl.SetExitKey(0)
}

func loadFont() rl.Font {
	font := rl.LoadFontEx("/usr/share/fonts/liberation/LiberationMono-Regular.ttf", 32, nil, 0)
	rl.SetTextureFilter(font.Texture, rl.FilterBilinear)
	return font
}

func NewApp(startModel string, interactive bool) *App {
	models := []string{
		"fluid", "nbody", "lorenz", "rossler", "pendulum", "double_pendulum", "cartpole",
		"spring_mass", "drone", "threebody", "coupled", "masschain", "gyroscope",
		"wave", "doublewell", "duffing", "magnetic", "vanderpol",
	}
	sort.Strings(models)

	app := &App{
		Models:   models,
		Selected: 0,
		Camera: rl.NewCamera3D(
			rl.NewVector3(0, 0, 50),
			rl.NewVector3(0, 0, 0),
			rl.NewVector3(0, 1, 0),
			45.0,
			rl.CameraPerspective,
		),
		Params:  make(map[string]float64),
		Font:    loadFont(),
		InMenu:  interactive,
		Running: !interactive,
	}

	if !interactive {
		app.loadModel(startModel)
	}

	return app
}

func RunInteractive() {
	initWindow()
	defer rl.CloseWindow()
	app := NewApp("", true)
	app.RunLoop()
}

func Run(modelName string) {
	initWindow()
	defer rl.CloseWindow()
	app := NewApp(modelName, false)
	app.RunLoop()
}

func (a *App) RunLoop() {
	for !rl.WindowShouldClose() {
		a.Update()
		a.Draw()
	}
}

func (a *App) loadModel(name string) {
	var dyn dynamo.System
	var state []float64

	switch name {
	case "fluid":
		dyn, state = physics.NewSPH(400), physics.NewSPH(400).DefaultState()
	case "nbody":
		dyn, state = physics.NewNBody(200), physics.NewNBody(200).DefaultState()
	case "lorenz":
		dyn, state = physics.NewLorenz(), []float64{1, 1, 1}
	case "rossler":
		dyn, state = physics.NewRossler(), []float64{1, 1, 1}
	case "doublewell":
		dyn, state = physics.NewDoubleWell(), []float64{0.5, 0}

	// Mechanical Systems
	case "pendulum":
		dyn, state = physics.NewPendulum(), []float64{0.5, 0}
	case "double_pendulum":
		dyn, state = physics.NewDoublePendulum(), []float64{0.5, 0.5, 0, 0}
	case "cartpole":
		dyn, state = physics.NewCartPole(), []float64{0, 0, 0.1, 0}
	case "spring_mass":
		dyn, state = physics.NewSpringMass(), []float64{1, 0}
	case "drone":
		dyn, state = physics.NewDrone(), []float64{0, 5, 0, 0, 0, 0}
	case "threebody":
		t := physics.NewThreeBody()
		dyn, state = t, t.DefaultState()
	case "coupled":
		dyn, state = physics.NewCoupledPendulums(), []float64{0.5, 0, 0, 0}
	case "masschain":
		m := physics.NewMassChain(40)
		dyn, state = m, m.DefaultState()
	case "gyroscope":
		g := physics.NewGyroscope()
		dyn, state = g, g.DefaultState()
	case "wave":
		w := physics.NewWave(50)
		dyn, state = w, w.DefaultState()
	case "duffing":
		d := physics.NewDuffing()
		dyn, state = d, d.DefaultState()
	case "magnetic":
		m := physics.NewMagneticPendulum()
		dyn, state = m, m.DefaultState()
	case "vanderpol":
		v := physics.NewVanDerPol()
		dyn, state = v, v.DefaultState()

	default:
		// Fallback to Pendulum if unknown
		dyn, state = physics.NewPendulum(), []float64{0.5, 0}
	}

	a.Dyn = dyn
	a.State = state
	a.Integ = integrators.NewRK4()
	a.Ctrl = control.NewNone(dyn.ControlDim())
	a.ModelName = name
	a.Time = 0
	a.Dt = 0.016
	a.Running = true
	// Only exit menu if we successfully loaded a model (which we essentially always do)
	// But if called from interactive menu, we set these flags:
	a.InMenu = false
	a.InConfig = false

	if cfg, ok := dyn.(dynamo.Configurable); ok {
		a.Params = cfg.GetParams()
	} else {
		a.Params = make(map[string]float64)
	}
	a.ParamKeys = make([]string, 0, len(a.Params))
	for k := range a.Params {
		a.ParamKeys = append(a.ParamKeys, k)
	}
	sort.Strings(a.ParamKeys)

	// Set Camera Defaults based on model type
	a.Camera = rl.NewCamera3D(
		rl.NewVector3(0, 0, 50),
		rl.NewVector3(0, 0, 0),
		rl.NewVector3(0, 1, 0),
		45.0,
		rl.CameraPerspective,
	)

	switch name {
	case "fluid":
		a.Camera.Position = rl.NewVector3(30, 20, 70)
		a.Camera.Target = rl.NewVector3(30, 20, 0)
	case "nbody":
		a.Camera.Position = rl.NewVector3(0, 0, 150)
	case "masschain", "wave":
		a.Camera.Position = rl.NewVector3(30, 10, 60)
		a.Camera.Target = rl.NewVector3(30, 10, 0)
	case "drone", "cartpole", "pendulum", "double_pendulum":
		a.Camera.Position = rl.NewVector3(0, 5, 20)
		a.Camera.Target = rl.NewVector3(0, 5, 0)
	}
}

func (a *App) Update() {
	if rl.IsKeyPressed(rl.KeyQ) {
		os.Exit(0)
	}

	if a.InMenu {
		if rl.IsKeyPressed(rl.KeyDown) || rl.IsKeyPressed(rl.KeyJ) {
			a.Selected++
		}
		if rl.IsKeyPressed(rl.KeyUp) || rl.IsKeyPressed(rl.KeyK) {
			a.Selected--
		}

		// Wrap selection
		if a.Selected >= len(a.Models) {
			a.Selected = 0
		}
		if a.Selected < 0 {
			a.Selected = len(a.Models) - 1
		}

		if rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeySpace) {
			a.loadModel(a.Models[a.Selected])
			a.InMenu = false
			a.InConfig = true // Go to config first
			a.Running = false
		}
		return
	}

	if a.InConfig {
		if rl.IsKeyPressed(rl.KeyEscape) {
			a.InMenu = true
			a.InConfig = false
			return
		}
		if rl.IsKeyPressed(rl.KeyEnter) {
			a.InConfig = false
			a.Running = true
			if cfg, ok := a.Dyn.(dynamo.Configurable); ok {
				for k, v := range a.Params {
					cfg.SetParam(k, v)
				}
			}
			return
		}

		if len(a.ParamKeys) > 0 {
			if rl.IsKeyPressed(rl.KeyDown) || rl.IsKeyPressed(rl.KeyJ) {
				a.ParamSel = (a.ParamSel + 1) % len(a.ParamKeys)
			}
			if rl.IsKeyPressed(rl.KeyUp) || rl.IsKeyPressed(rl.KeyK) {
				// Go backwards correctly
				a.ParamSel--
				if a.ParamSel < 0 {
					a.ParamSel = len(a.ParamKeys) - 1
				}
			}

			key := a.ParamKeys[a.ParamSel]
			step := 0.1
			if rl.IsKeyDown(rl.KeyLeftShift) {
				step = 1.0
			}

			if rl.IsKeyPressed(rl.KeyRight) || rl.IsKeyPressed(rl.KeyL) {
				a.Params[key] += step
			}
			if rl.IsKeyPressed(rl.KeyLeft) || rl.IsKeyPressed(rl.KeyH) {
				a.Params[key] -= step
			}
		}
		return
	}

	// Simulation Running or Paused
	if rl.IsKeyPressed(rl.KeyEscape) {
		a.InMenu = true
		a.Running = false
		return
	}

	if a.Running {
		u := a.Ctrl.Compute(a.State, a.Time)
		a.State = a.Integ.Step(a.Dyn, a.State, u, a.Time, a.Dt)
		a.Time += a.Dt
	}

	// Camera & Interaction
	if rl.IsKeyDown(rl.KeyW) {
		a.Camera.Position.Y += 0.5
	}
	if rl.IsKeyDown(rl.KeyS) {
		a.Camera.Position.Y -= 0.5
	}
	if rl.IsKeyDown(rl.KeyA) {
		a.Camera.Position.X -= 0.5
	}
	if rl.IsKeyDown(rl.KeyD) {
		a.Camera.Position.X += 0.5
	}

	if rl.IsMouseButtonDown(rl.MouseRightButton) {
		delta := rl.GetMouseDelta()
		a.Camera.Position.X -= delta.X * 0.2
		a.Camera.Position.Y += delta.Y * 0.2
	}

	// Zoom
	wheel := rl.GetMouseWheelMove()
	if wheel != 0 {
		zoom := float32(wheel) * 3.0
		diff := rl.Vector3Subtract(a.Camera.Target, a.Camera.Position)
		dist := rl.Vector3Length(diff)
		if dist > 5.0 || zoom < 0 {
			dir := rl.Vector3Normalize(diff)
			a.Camera.Position = rl.Vector3Add(a.Camera.Position, rl.Vector3Scale(dir, zoom))
		}
	}

	if rl.IsKeyPressed(rl.KeySpace) {
		a.Running = !a.Running
	}
	if rl.IsKeyPressed(rl.KeyR) {
		// Reset simulation
		currentParams := a.Params
		a.loadModel(a.ModelName)
		a.Params = currentParams
		a.InConfig = false
		a.Running = true
		if cfg, ok := a.Dyn.(dynamo.Configurable); ok {
			for k, v := range a.Params {
				cfg.SetParam(k, v)
			}
		}
	}
}

func (a *App) Draw() {
	rl.BeginDrawing()
	rl.ClearBackground(ColBg)

	if a.InMenu {
		a.drawMenu()
	} else if a.InConfig {
		a.drawConfig()
	} else {
		a.drawSim()
	}
	rl.EndDrawing()
}

func (a *App) drawText(text string, x, y int, size int, color rl.Color) {
	rl.DrawTextEx(a.Font, text, rl.NewVector2(float32(x), float32(y)), float32(size), 1, color)
}

func (a *App) CustomGrid(slices int, spacing float32) {
	halfSize := float32(slices) * spacing / 2
	rl.BeginMode3D(a.Camera)
	// Grid Lines
	for i := -slices / 2; i <= slices/2; i++ {
		pos := float32(i) * spacing
		rl.DrawLine3D(rl.NewVector3(pos, 0, -halfSize), rl.NewVector3(pos, 0, halfSize), ColGrid)
		rl.DrawLine3D(rl.NewVector3(-halfSize, 0, pos), rl.NewVector3(halfSize, 0, pos), ColGrid)
	}
	rl.EndMode3D()
}

func (a *App) drawSim() {
	a.CustomGrid(60, 5.0)
	rl.BeginMode3D(a.Camera)
	switch a.ModelName {
	case "fluid":
		a.RenderSPH()
	case "nbody":
		a.RenderNBody()
	case "lorenz", "rossler":
		a.RenderAttractor()
	case "doublewell":
		a.RenderDoubleWell()
	case "duffing":
		a.RenderDuffing()
	case "vanderpol":
		a.RenderAttractor()
	case "pendulum":
		a.RenderPendulum()
	case "double_pendulum":
		a.RenderDoublePendulum()
	case "cartpole":
		a.RenderCartPole()
	case "spring_mass":
		a.RenderSpringMass()
	case "drone":
		a.RenderDrone()
	case "threebody":
		a.RenderThreeBody()
	case "coupled":
		a.RenderCoupled()
	case "masschain":
		a.RenderMassChain()
	case "gyroscope":
		a.RenderGyroscope()
	case "wave":
		a.RenderWave()
	case "magnetic":
		a.RenderMagnetic()
	default:
		a.RenderGeneric()
	}
	rl.EndMode3D()

	// HUD
	a.drawText("dynsim", 30, 30, 24, ColSelect)
	a.drawText(fmt.Sprintf(":: %s", a.ModelName), 140, 34, 16, ColText)

	status := "RUNNING"
	col := ColSelect
	if !a.Running {
		status = "PAUSED"
		col = ColTextDim
	}
	a.drawText(status, 1150, 30, 16, col)

	// Bottom Controls
	a.drawText("[SPACE] PAUSE  [R] RESET  [ESC] MENU  [Q] QUIT", 820, 680, 14, ColTextDim)
	a.drawText(fmt.Sprintf("%d FPS", int32(rl.GetFPS())), 30, 680, 14, ColTextDim)
}

func (a *App) drawMenu() {
	a.drawText("dynsim", 50, 50, 40, ColSelect)
	a.drawText("Select Simulation", 50, 100, 16, ColTextDim)

	limit := 18
	startIdx := 0
	if a.Selected >= limit {
		startIdx = a.Selected - limit + 1
	}

	y := 160
	for i := startIdx; i < len(a.Models) && i < startIdx+limit; i++ {
		name := a.Models[i]
		isSel := (i == a.Selected)
		if isSel {
			a.drawText(fmt.Sprintf("> %s", name), 50, y, 20, ColSelect)
		} else {
			a.drawText(fmt.Sprintf("  %s", name), 50, y, 20, ColText)
		}
		y += 28
	}

	a.drawText("ARROWS: NAVIGATE  ENTER: SELECT  Q: QUIT", 850, 680, 14, ColTextDim)
}

func (a *App) drawConfig() {
	a.drawText("dynsim", 50, 50, 40, ColTextDim)
	a.drawText("configure", 220, 65, 20, ColSelect)
	a.drawText(fmt.Sprintf("Target: %s", a.ModelName), 50, 110, 16, ColAccent)

	y := 180
	if len(a.ParamKeys) == 0 {
		a.drawText("No configurable parameters.", 50, y, 16, ColTextDim)
	} else {
		for i, key := range a.ParamKeys {
			isSel := (i == a.ParamSel)
			val := a.Params[key]
			if isSel {
				a.drawText(fmt.Sprintf("> %-15s %.2f", key, val), 50, y, 20, ColSelect)
			} else {
				a.drawText(fmt.Sprintf("  %-15s %.2f", key, val), 50, y, 20, ColText)
			}
			y += 28
		}
	}

	a.drawText("ARROWS: ADJUST  ENTER: RUN  ESC: BACK", 880, 680, 14, ColTextDim)
}
