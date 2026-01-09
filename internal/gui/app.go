package gui

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/san-kum/dynsim/internal/audio"
	"github.com/san-kum/dynsim/internal/compute" // New import
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
	Dyn         dynamo.System
	Integ       dynamo.Integrator
	Ctrl        dynamo.Controller
	State       dynamo.State
	ModelName   string
	Time        float64
	Dt          float64
	Camera      rl.Camera3D
	Running     bool
	InMenu      bool
	InConfig    bool
	Models      []string
	Selected    int
	Params      map[string]float64
	ParamKeys   []string
	ParamSel    int
	History     [][]float64 // Ring buffer for trails
	Telemetry   []float64   // Ring buffer for scalar graph (e.g. velocity mag)
	MaxHistory  int
	ShowVectors bool
	Font        rl.Font

	// Visual Polish
	ParticleTex  rl.Texture2D
	CamPosTarget rl.Vector3
	CamTgtTarget rl.Vector3
	CursorViz    rl.Vector3 // [x, y, active(1.0/0.0)]
	Stars        []float64  // Background stars [x, y, z]

	// Post-Processing
	TargetTex   rl.RenderTexture2D
	BloomShader rl.Shader

	// Audio
	Audio *audio.Processor

	// Compute
	UseCompute bool
	GLBackend  *compute.OpenGLBackend
}

// initWindow initializes the Raylib window with size 1280×720 and title "dynsim", sets the target FPS to 60, and disables the default exit key.
func initWindow() {
	rl.InitWindow(1280, 720, "dynsim")
	rl.SetTargetFPS(60)
	rl.SetExitKey(0)
}

// loadFont loads the Liberation Mono font from the system path and enables bilinear texture filtering.
// It returns the loaded rl.Font ready for use in rendering.
func loadFont() rl.Font {
	font := rl.LoadFontEx("/usr/share/fonts/liberation/LiberationMono-Regular.ttf", 32, nil, 0)
	rl.SetTextureFilter(font.Texture, rl.FilterBilinear)
	return font
}

// NewApp creates and initializes an App configured for either interactive menu-driven use or direct model execution.
// If interactive is false, the provided startModel is loaded immediately; otherwise the app starts in the model selection menu.
// It returns a pointer to the initialized App.
func NewApp(startModel string, interactive bool) *App {
	models := []string{
		"fluid", "nbody", "lorenz", "rossler", "pendulum", "double_pendulum", "cartpole",
		"spring_mass", "drone", "threebody", "coupled", "masschain", "gyroscope",
		"wave", "doublewell", "duffing", "magnetic", "vanderpol",
	}
	sort.Strings(models)

	// Audio Init
	proc := audio.NewProcessor()
	// proc.Start() // Output-only mode is active by default in struct but start logic is commented for now to avoid conflicts if previously failed.
	// Actually let's assume Audio is working now.
	proc.Start()

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
		Params:      make(map[string]float64),
		Font:        loadFont(),
		InMenu:      interactive,
		Running:     !interactive,
		MaxHistory:  50, // Store last 50 frames
		History:     make([][]float64, 0, 50),
		Telemetry:   make([]float64, 0, 200), // Longer history for graphs
		ShowVectors: true,
		Audio:       proc,
	}

	// Generate Glow Texture
	img := rl.GenImageGradientRadial(32, 32, 0.0, rl.White, rl.NewColor(0, 0, 0, 0))
	app.ParticleTex = rl.LoadTextureFromImage(img)
	rl.UnloadImage(img)

	// Generate Starfield
	numStars := 2000
	app.Stars = make([]float64, numStars*3)
	for i := 0; i < numStars; i++ {
		// Wide spread, deep background
		app.Stars[i*3] = (rand.Float64() - 0.5) * 1000
		app.Stars[i*3+1] = (rand.Float64() - 0.5) * 1000
		app.Stars[i*3+2] = -500 - rand.Float64()*500 // Behind everything
	}

	// Init Post Processing
	app.TargetTex = rl.LoadRenderTexture(1280, 720)
	app.BloomShader = rl.LoadShader("", "assets/shaders/bloom.fs")

	if !interactive {
		app.loadModel(startModel)
	}

	return app
}

// RunInteractive initializes the graphical window, creates an interactive App, and enters its main run loop.
// It blocks until the window is closed and ensures Raylib's window is closed on return.
func RunInteractive() {
	initWindow()
	defer rl.CloseWindow()
	app := NewApp("", true)
	app.RunLoop()
}

// Run starts a non-interactive GUI session for the specified model and enters the main update-draw loop.
// It initializes the window, loads modelName into the application, and blocks until the window is closed.
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

	// Reset Backend Flags
	a.UseCompute = false

	// Default Camera
	a.Camera = rl.NewCamera3D(
		rl.NewVector3(0, 0, 50),
		rl.NewVector3(0, 0, 0),
		rl.NewVector3(0, 1, 0),
		45.0,
		rl.CameraPerspective,
	)

	switch name {
	case "fluid":
		dyn, state = physics.NewSPH(400), physics.NewSPH(400).DefaultState()
		a.Camera.Position = rl.NewVector3(30, 20, 70)
		a.Camera.Target = rl.NewVector3(30, 20, 0)
	case "nbody":
		// Massive Scale GPU Implementation
		// We don't use CPU dyn/state for physics, just for controller placeholder
		dyn, state = physics.NewNBody(1), []float64{0}
		a.UseCompute = true
		a.GLBackend = compute.NewOpenGLBackend(65536) // Start with 64k particles
		a.Camera.Position = rl.NewVector3(0, 0, 400)

	case "hybrid":
		// 8k Stars + 4k Gas
		h := physics.NewHybrid(8192, 4096)
		dyn, state = h, h.DefaultState()
		a.Camera.Position = rl.NewVector3(0, 0, 150)
	case "lorenz":
		dyn, state = physics.NewLorenz(), []float64{1, 1, 1}
	case "rossler":
		dyn, state = physics.NewRossler(), []float64{1, 1, 1}
	case "doublewell":
		dyn, state = physics.NewDoubleWell(), []float64{0.5, 0}

	// Mechanical Systems
	case "pendulum":
		dyn, state = physics.NewPendulum(), []float64{0.5, 0}
		a.Camera.Position = rl.NewVector3(0, 5, 20)
		a.Camera.Target = rl.NewVector3(0, 5, 0)
	case "double_pendulum":
		dyn, state = physics.NewDoublePendulum(), []float64{0.5, 0.5, 0, 0}
		a.Camera.Position = rl.NewVector3(0, 5, 20)
		a.Camera.Target = rl.NewVector3(0, 5, 0)
	case "cartpole":
		dyn, state = physics.NewCartPole(), []float64{0, 0, 0.1, 0}
		a.Camera.Position = rl.NewVector3(0, 5, 20)
		a.Camera.Target = rl.NewVector3(0, 5, 0)
	case "spring_mass":
		dyn, state = physics.NewSpringMass(), []float64{1, 0}
	case "drone":
		dyn, state = physics.NewDrone(), []float64{0, 5, 0, 0, 0, 0}
		a.Camera.Position = rl.NewVector3(0, 5, 20)
		a.Camera.Target = rl.NewVector3(0, 5, 0)
	case "threebody":
		t := physics.NewThreeBody()
		dyn, state = t, t.DefaultState()
	case "coupled":
		dyn, state = physics.NewCoupledPendulums(), []float64{0.5, 0, 0, 0}
	case "masschain":
		m := physics.NewMassChain(40)
		dyn, state = m, m.DefaultState()
		a.Camera.Position = rl.NewVector3(30, 10, 60)
		a.Camera.Target = rl.NewVector3(30, 10, 0)
	case "gyroscope":
		g := physics.NewGyroscope()
		dyn, state = g, g.DefaultState()
	case "wave":
		w := physics.NewWave(50)
		dyn, state = w, w.DefaultState()
		a.Camera.Position = rl.NewVector3(30, 10, 60)
		a.Camera.Target = rl.NewVector3(30, 10, 0)
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

	// Controller Selection
	ctrlDim := dyn.ControlDim()
	if ctrlDim == 3 {
		// Enable Hand of God
		a.Ctrl = control.NewManual()
	} else {
		a.Ctrl = control.NewNone(ctrlDim)
	}

	a.ModelName = name
	a.Time = 0
	a.Dt = 0.016
	a.Running = true
	a.History = make([][]float64, 0, a.MaxHistory) // Clear trails
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

	// Initialize Smooth Camera Targets
	a.CamPosTarget = a.Camera.Position
	a.CamTgtTarget = a.Camera.Target
}

func (a *App) Update() {
	if rl.IsKeyPressed(rl.KeyQ) {
		a.Audio.Stop()
		os.Exit(0)
	}

	// Update Audio
	a.Audio.Update()

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

	// Interaction (Hand of God) - Calc before Sim
	// Raycast Mouse to World Z=0 Plane
	mousePos := rl.GetMousePosition()
	ray := rl.GetMouseRay(mousePos, a.Camera)

	worldX, worldY, mouseStrength := 0.0, 0.0, 0.0

	if ray.Direction.Z != 0 {
		t := -ray.Position.Z / ray.Direction.Z
		if t > 0 {
			worldX = float64(ray.Position.X + t*ray.Direction.X)
			worldY = float64(ray.Position.Y + t*ray.Direction.Y)

			if rl.IsMouseButtonDown(rl.MouseLeftButton) {
				mouseStrength = 50.0 // Attract
			} else if rl.IsMouseButtonDown(rl.MouseRightButton) {
				mouseStrength = -50.0 // Repel
			}

			// Visual Cursor
			if mouseStrength != 0 {
				a.CursorViz = rl.NewVector3(float32(worldX), float32(worldY), 1.0)
			} else {
				a.CursorViz.Z = 0 // Inactive
			}
		}
	} else {
		a.CursorViz.Z = 0
	}

	// Update Manual Controller (CPU)
	if man, ok := a.Ctrl.(*control.ManualController); ok {
		man.SetControl([]float64{worldX, worldY, mouseStrength})
	}

	// COMPUTE SHADER MODE
	if a.Running && a.UseCompute {
		// Initialize if first run
		if !a.GLBackend.Initialized {
			// Convert current state to float32 buffer
			N := int(a.GLBackend.NumParticles)
			data := make([]float32, N*8)
			for i := 0; i < N; i++ {
				// Initialize random Galaxy
				r := 10.0 + rand.Float64()*100.0
				theta := rand.Float64() * 6.28
				height := (rand.Float64() - 0.5) * 10.0

				Px := r * math.Cos(theta)
				Py := r * math.Sin(theta)
				Pz := height

				// Orbital Velocity
				v := math.Sqrt(100000.0 / r)
				Vx := -v * math.Sin(theta)
				Vy := v * math.Cos(theta)

				data[i*8+0] = float32(Px)
				data[i*8+1] = float32(Py)
				data[i*8+2] = float32(Pz)
				data[i*8+3] = 1.0 // Mass

				data[i*8+4] = float32(Vx)
				data[i*8+5] = float32(Vy)
				data[i*8+6] = 0.0
				data[i*8+7] = 0.0 // Pad
			}

			err := a.GLBackend.Init("assets/shaders/nbody.comp", data)
			if err != nil {
				fmt.Printf("Compute Init Error: %v\n", err)
				a.UseCompute = false // Fallback
			} else {
				// Init Render Shader
				if err := a.GLBackend.InitRender("assets/shaders/render.vert", "assets/shaders/render.frag"); err != nil {
					fmt.Printf("Render Init Error: %v\n", err)
				}
				a.Camera.Position = rl.NewVector3(0, 0, 300) // Zoom out for massive scale
			}
		} else {
			// Step Compute
			// Pass Mouse Data as float32
			a.GLBackend.Step(0.016, 100.0, float32(worldX), float32(worldY), float32(mouseStrength))
		}

		// Audio from Compute? We need to download data (slow) or just fake it based on camera?
		// For massive scale, let's just drive audio with a constant humble-brag drone.
		if a.Audio != nil {
			a.Audio.UpdatePhysics(50000.0, 10.0, 0.0) // Fake energy
		}

		return // Skip CPU Logic
	}

	if a.Running {
		u := a.Ctrl.Compute(a.State, a.Time)

		// Inject Audio into Control Vector if Interactive
		if a.ModelName == "hybrid" || a.ModelName == "fluid" || a.Dyn.ControlDim() >= 3 {
			// Ensure u has space
			// ManualController returns 3
			u = append(u, a.Audio.Bass, a.Audio.Mid, a.Audio.High)
		}

		a.State = a.Integ.Step(a.Dyn, a.State, u, a.Time, a.Dt)
		a.Time += a.Dt

		// Record History & Telemetry
		stateCopy := make([]float64, len(a.State))
		copy(stateCopy, a.State)
		a.History = append(a.History, stateCopy)
		if len(a.History) > a.MaxHistory {
			a.History = a.History[1:]
		}

		// Telemetry
		var metric float64
		if len(u) > 0 {
			sum := 0.0
			for _, v := range a.State {
				sum += v * v
			}
			metric = sum / float64(len(a.State))
		}
		a.Telemetry = append(a.Telemetry, metric)
		if len(a.Telemetry) > 200 {
			a.Telemetry = a.Telemetry[1:]
		}

		// Sonification
		if a.Audio != nil && a.Audio.Active {
			// Physics -> Audio Mapping
			// Use metric (Avg Velocity/Energy) to open the Filter
			// Scale metric carefully
			a.Audio.UpdatePhysics(metric*500.0, 0, 0)
		}
	}

	// Interaction (Hand of God)

	// Camera & Interaction
	if rl.IsKeyPressed(rl.KeyV) {
		a.ShowVectors = !a.ShowVectors
	}

	// Input modifies the TARGET, not the camera directly
	if rl.IsKeyDown(rl.KeyW) {
		a.CamPosTarget.Y += 0.5
	}
	if rl.IsKeyDown(rl.KeyS) {
		a.CamPosTarget.Y -= 0.5
	}
	if rl.IsKeyDown(rl.KeyA) {
		a.CamPosTarget.X -= 0.5
	}
	if rl.IsKeyDown(rl.KeyD) {
		a.CamPosTarget.X += 0.5
	}

	if rl.IsMouseButtonDown(rl.MouseRightButton) {
		delta := rl.GetMouseDelta()
		a.CamPosTarget.X -= delta.X * 0.2
		a.CamPosTarget.Y += delta.Y * 0.2
	}

	// Zoom
	wheel := rl.GetMouseWheelMove()
	if wheel != 0 {
		zoom := float32(wheel) * 3.0
		diff := rl.Vector3Subtract(a.CamTgtTarget, a.CamPosTarget)
		dist := rl.Vector3Length(diff)
		if dist > 5.0 || zoom < 0 {
			dir := rl.Vector3Normalize(diff)
			a.CamPosTarget = rl.Vector3Add(a.CamPosTarget, rl.Vector3Scale(dir, zoom))
		}
	}

	// Apply Inertia (Lerp)
	lerp := float32(5.0 * a.Dt)
	if lerp > 1.0 {
		lerp = 1.0
	}

	a.Camera.Position = rl.Vector3Lerp(a.Camera.Position, a.CamPosTarget, lerp)
	a.Camera.Target = rl.Vector3Lerp(a.Camera.Target, a.CamTgtTarget, lerp)

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
		a.DrawHUD()
	}

	rl.EndDrawing()
}

func (a *App) DrawHUD() {
	a.drawText("dynsim", 30, 30, 24, ColSelect)
	a.drawText(fmt.Sprintf(":: %s", a.ModelName), 140, 34, 16, ColText)

	a.DrawTelemetry()

	status := "RUNNING"
	col := ColSelect
	if !a.Running {
		status = "PAUSED"
		col = ColTextDim
	}
	a.drawText(status, 1150, 30, 16, col)

	a.drawText("[SPACE] PAUSE  [R] RESET  [V] VECTORS  [ESC] MENU  [Q] QUIT", 700, 680, 14, ColTextDim)
	a.drawText(fmt.Sprintf("%d FPS", int32(rl.GetFPS())), 30, 680, 14, ColTextDim)

	// Mic Level
	if a.Audio != nil && a.Audio.Active {
		sum := (a.Audio.Bass + a.Audio.Mid + a.Audio.High) / 3.0
		bars := int(sum * 20)
		if bars > 20 {
			bars = 20
		}
		barStr := ""
		for i := 0; i < bars; i++ {
			barStr += "|"
		}
		a.drawText(fmt.Sprintf("MIC [%-20s]", barStr), 30, 650, 14, ColAccent)
	} else {
		a.drawText("MIC [OFF]", 30, 650, 14, rl.Red)
	}
}

// DrawScanlines ...
func (a *App) DrawScanlines() {
	// Draw alternating faint lines
	h := int32(720)
	for y := int32(0); y < h; y += 4 {
		rl.DrawRectangle(0, y, 1280, 2, rl.NewColor(0, 0, 0, 40))
	}
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
	// a.CustomGrid(60, 5.0)
	rl.BeginMode3D(a.Camera)
	switch a.ModelName {
	case "fluid":
		a.RenderSPH()
	case "nbody":
		a.RenderComputeNBody() // Replaced CPU Render
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

	// Draw Interaction Cursor
	if a.CursorViz.Z > 0 {
		pos := rl.NewVector3(a.CursorViz.X, a.CursorViz.Y, 0)
		rl.DrawCircle3D(pos, 2.0, rl.NewVector3(0, 0, 1), 0, rl.NewColor(255, 255, 255, 100))
		rl.DrawCircle3D(pos, 5.0, rl.NewVector3(0, 0, 1), 0, rl.NewColor(255, 255, 255, 50))
	}

	rl.EndMode3D()
}

func (a *App) DrawTelemetry() {
	if len(a.Telemetry) < 2 {
		return
	}

	rectX, rectY := 30, 600
	width, height := 400, 60

	// Normalize Data
	minVal, maxVal := a.Telemetry[0], a.Telemetry[0]
	for _, v := range a.Telemetry {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal == minVal {
		maxVal = minVal + 1
	}

	// Draw Line Strip
	points := make([]rl.Vector2, len(a.Telemetry))
	for i, val := range a.Telemetry {
		px := float32(rectX) + (float32(i)/float32(len(a.Telemetry)))*float32(width)
		norm := (val - minVal) / (maxVal - minVal)
		py := float32(rectY+height) - float32(norm)*float32(height)
		points[i] = rl.NewVector2(px, py)
	}

	rl.DrawLineStrip(points, ColAccent)
	a.drawText(fmt.Sprintf("E: %.2e", a.Telemetry[len(a.Telemetry)-1]), rectX+width+10, rectY+height-10, 14, ColText)
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