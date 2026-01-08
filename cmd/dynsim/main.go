package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/guptarohit/asciigraph"
	"github.com/san-kum/dynsim/internal/analysis"
	"github.com/san-kum/dynsim/internal/config"
	"github.com/san-kum/dynsim/internal/control"
	"github.com/san-kum/dynsim/internal/dynamo"
	"github.com/san-kum/dynsim/internal/experiment"
	"github.com/san-kum/dynsim/internal/gui"
	"github.com/san-kum/dynsim/internal/storage"
	"github.com/san-kum/dynsim/internal/viz"
	"github.com/spf13/cobra"
)

var (
	dataDir    string
	dt         float64
	duration   float64
	theta      float64
	omega      float64
	pos        float64
	vel        float64
	seed       int64
	integrator string
	controller string
	kp         float64
	ki         float64
	kd         float64
	target     float64
	numBodies  int
	// New model parameters
	theta2  float64 // double pendulum second angle
	omega2  float64 // double pendulum second angular velocity
	thrustL float64 // drone left thrust
	thrustR float64 // drone right thrust
	// Phase plot axes
	xAxis int
	yAxis int
	// Config file
	configFile string
	// Frame rate for live view
	frameRate int
	// Preset name
	preset string
)

// main is the entry point for the dynsim CLI; it registers commands and flags, launches the interactive GUI when no subcommand is provided, and executes the root command.
// It exits the process with status 1 if command execution returns an error.
func main() {
	rootCmd := &cobra.Command{
		Use:   "dynsim",
		Short: "physics and control simulation lab",
		Run: func(cmd *cobra.Command, args []string) {
			// Default to interactive GUI mode when no command given
			gui.RunInteractive()
		},
	}

	rootCmd.PersistentFlags().StringVar(&dataDir, "data", ".dynsim", "data directory")

	runCmd := &cobra.Command{
		Use:   "run [model]",
		Short: "run simulation",
		Args:  cobra.ExactArgs(1),
		RunE:  runSimulation,
	}
	runCmd.Flags().Float64Var(&dt, "dt", 0.01, "timestep")
	runCmd.Flags().Float64Var(&duration, "time", 10.0, "duration")
	runCmd.Flags().Float64Var(&theta, "theta", 0.5, "initial angle")
	runCmd.Flags().Float64Var(&omega, "omega", 0.0, "initial angular velocity")
	runCmd.Flags().Float64Var(&pos, "pos", 0.0, "initial position (cartpole)")
	runCmd.Flags().Float64Var(&vel, "vel", 0.0, "initial velocity (cartpole)")
	runCmd.Flags().Int64Var(&seed, "seed", time.Now().UnixNano(), "random seed")
	runCmd.Flags().StringVar(&integrator, "integrator", "rk4", "integrator")
	runCmd.Flags().StringVar(&controller, "controller", "none", "controller")
	runCmd.Flags().Float64Var(&kp, "kp", 10.0, "pid kp")
	runCmd.Flags().Float64Var(&ki, "ki", 0.1, "pid ki")
	runCmd.Flags().Float64Var(&kd, "kd", 5.0, "pid kd")
	runCmd.Flags().Float64Var(&target, "target", 0.0, "pid target")
	runCmd.Flags().IntVar(&numBodies, "bodies", 3, "number of bodies (nbody)")
	runCmd.Flags().Float64Var(&theta2, "theta2", 0.5, "second angle (double_pendulum)")
	runCmd.Flags().Float64Var(&omega2, "omega2", 0.0, "second angular velocity (double_pendulum)")
	runCmd.Flags().StringVar(&configFile, "config", "", "config file path (yaml)")
	runCmd.Flags().StringVar(&preset, "preset", "", "use preset configuration")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "list runs",
		RunE:  listRuns,
	}

	plotCmd := &cobra.Command{
		Use:   "plot [run_id]",
		Short: "plot run results",
		Args:  cobra.ExactArgs(1),
		RunE:  plotRun,
	}

	exportCmd := &cobra.Command{
		Use:   "export [run_id]",
		Short: "export run metadata",
		Args:  cobra.ExactArgs(1),
		RunE:  exportRun,
	}

	benchCmd := &cobra.Command{
		Use:   "bench [model]",
		Short: "benchmark model",
		Args:  cobra.ExactArgs(1),
		RunE:  benchModel,
	}

	analyzeCmd := &cobra.Command{
		Use:   "analyze [run_id]",
		Short: "frequency analysis",
		Args:  cobra.ExactArgs(1),
		RunE:  analyzeRun,
	}

	// New commands
	liveCmd := &cobra.Command{
		Use:   "live [model]",
		Short: "run simulation with live visualization",
		Args:  cobra.ExactArgs(1),
		RunE:  runLive,
	}
	liveCmd.Flags().Float64Var(&dt, "dt", 0.01, "timestep")
	liveCmd.Flags().Float64Var(&duration, "time", 10.0, "duration")
	liveCmd.Flags().Float64Var(&theta, "theta", 0.5, "initial angle")
	liveCmd.Flags().Float64Var(&omega, "omega", 0.0, "initial angular velocity")
	liveCmd.Flags().Float64Var(&pos, "pos", 0.0, "initial position")
	liveCmd.Flags().Float64Var(&vel, "vel", 0.0, "initial velocity")
	liveCmd.Flags().Float64Var(&theta2, "theta2", 0.5, "second angle (double_pendulum)")
	liveCmd.Flags().Float64Var(&omega2, "omega2", 0.0, "second angular velocity")
	liveCmd.Flags().StringVar(&integrator, "integrator", "rk4", "integrator")
	liveCmd.Flags().StringVar(&controller, "controller", "none", "controller")
	liveCmd.Flags().Float64Var(&kp, "kp", 10.0, "pid kp")
	liveCmd.Flags().Float64Var(&ki, "ki", 0.1, "pid ki")
	liveCmd.Flags().Float64Var(&kd, "kd", 5.0, "pid kd")
	liveCmd.Flags().Float64Var(&target, "target", 0.0, "pid target")
	liveCmd.Flags().IntVar(&frameRate, "fps", 30, "frame rate")

	phaseCmd := &cobra.Command{
		Use:   "phase [run_id]",
		Short: "phase space plot",
		Args:  cobra.ExactArgs(1),
		RunE:  phasePlot,
	}
	phaseCmd.Flags().IntVar(&xAxis, "x-axis", 0, "state index for x-axis")
	phaseCmd.Flags().IntVar(&yAxis, "y-axis", 1, "state index for y-axis")

	exportCSVCmd := &cobra.Command{
		Use:   "export-csv [run_id]",
		Short: "export run data to CSV",
		Args:  cobra.ExactArgs(1),
		RunE:  exportCSV,
	}

	tuiCmd := &cobra.Command{
		Use:   "tui",
		Short: "legacy terminal TUI mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			return viz.RunInteractive()
		},
	}

	compareCmd := &cobra.Command{
		Use:   "compare [model] [integrator1] [integrator2] ...",
		Short: "compare integrators on the same model",
		Args:  cobra.MinimumNArgs(2),
		RunE:  compareIntegrators,
	}
	compareCmd.Flags().Float64Var(&dt, "dt", 0.01, "timestep")
	compareCmd.Flags().Float64Var(&duration, "time", 10.0, "duration")
	compareCmd.Flags().Float64Var(&theta, "theta", 0.5, "initial angle")

	presetsCmd := &cobra.Command{
		Use:   "presets [model]",
		Short: "list available presets for a model",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			presets := config.ListPresets(args[0])
			if len(presets) == 0 {
				fmt.Printf("no presets for model: %s\n", args[0])
				return nil
			}
			fmt.Printf("presets for %s:\n", args[0])
			for _, p := range presets {
				fmt.Printf("  %s\n", p)
			}
			return nil
		},
	}

	exportJSONCmd := &cobra.Command{
		Use:   "export-json [run_id]",
		Short: "export run data to JSON",
		Args:  cobra.ExactArgs(1),
		RunE:  exportJSON,
	}

	guiCmd := &cobra.Command{
		Use:   "gui [model]",
		Short: "run simulation with high-performance GUI",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			model := "fluid"
			if len(args) > 0 {
				model = args[0]
			}
			gui.Run(model)
		},
	}

	rootCmd.AddCommand(runCmd, listCmd, plotCmd, exportCmd, benchCmd, analyzeCmd, liveCmd, phaseCmd, exportCSVCmd, tuiCmd, compareCmd, presetsCmd, exportJSONCmd, guiCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
func runSimulation(cmd *cobra.Command, args []string) error {
	model := args[0]

	// Load preset if specified
	if preset != "" {
		cfg := config.GetPreset(model, preset)
		if cfg == nil {
			return fmt.Errorf("unknown preset: %s (available: %v)", preset, config.ListPresets(model))
		}
		// Apply preset values
		dt = cfg.Dt
		duration = cfg.Duration
		integrator = cfg.Integrator
		if cfg.Controller != "" {
			controller = cfg.Controller
		}
		theta = cfg.InitState.Theta
		omega = cfg.InitState.Omega
		theta2 = cfg.InitState.Theta2
		omega2 = cfg.InitState.Omega2
		pos = cfg.InitState.Pos
		vel = cfg.InitState.Vel
	}

	// Load config file if specified (overrides preset)
	if configFile != "" {
		cfg, err := config.Load(configFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		// Apply config values (CLI flags override config)
		if !cmd.Flags().Changed("dt") {
			dt = cfg.Dt
		}
		if !cmd.Flags().Changed("time") {
			duration = cfg.Duration
		}
		if !cmd.Flags().Changed("integrator") {
			integrator = cfg.Integrator
		}
		if !cmd.Flags().Changed("controller") {
			controller = cfg.Controller
		}
		if !cmd.Flags().Changed("theta") {
			theta = cfg.InitState.Theta
		}
		if !cmd.Flags().Changed("omega") {
			omega = cfg.InitState.Omega
		}
		if !cmd.Flags().Changed("pos") {
			pos = cfg.InitState.Pos
		}
		if !cmd.Flags().Changed("vel") {
			vel = cfg.InitState.Vel
		}
		if !cmd.Flags().Changed("kp") {
			kp = cfg.ControllerParams.Kp
		}
		if !cmd.Flags().Changed("ki") {
			ki = cfg.ControllerParams.Ki
		}
		if !cmd.Flags().Changed("kd") {
			kd = cfg.ControllerParams.Kd
		}
		if !cmd.Flags().Changed("target") {
			target = cfg.ControllerParams.Target
		}
		if cfg.Seed != 0 && !cmd.Flags().Changed("seed") {
			seed = cfg.Seed
		}
	}

	st := storage.New(dataDir)
	if err := st.Init(); err != nil {
		return err
	}

	registry := experiment.NewRegistry()

	dyn, err := registry.GetModel(model)
	if err != nil {
		return err
	}

	integ, err := registry.GetIntegrator(integrator)
	if err != nil {
		return err
	}

	controllerParams := map[string]float64{
		"dim":    float64(dyn.ControlDim()),
		"kp":     kp,
		"ki":     ki,
		"kd":     kd,
		"target": target,
	}
	ctrl, err := registry.GetController(controller, controllerParams)
	if err != nil {
		return err
	}

	var initState []float64
	switch model {
	case "cartpole":
		initState = []float64{pos, vel, theta, omega}
	case "nbody":
		initState = makeNBodyInitialState(numBodies)
	case "double_pendulum":
		initState = []float64{theta, theta2, omega, omega2}
	case "spring_mass":
		initState = []float64{pos, vel}
	case "spring_chain":
		initState = []float64{pos, 0, 0, vel, 0, 0} // 3 masses
	case "drone":
		initState = []float64{0, 5, theta, 0, 0, omega} // x, y, theta, vx, vy, omega
	default:
		initState = []float64{theta, omega}
	}

	cfg := experiment.Config{
		Model:      model,
		Integrator: integrator,
		Controller: controller,
		InitState:  initState,
		Dt:         dt,
		Duration:   duration,
		Seed:       seed,
		Params:     controllerParams,
	}

	exp := experiment.New(cfg)
	metrics := registry.DefaultMetrics(model)
	if err := exp.Setup(dyn, integ, ctrl, metrics); err != nil {
		return err
	}

	fmt.Printf("running %s simulation...\n", model)
	start := time.Now()

	result, err := exp.Run(context.Background())
	if err != nil {
		return err
	}

	elapsed := time.Since(start)

	runID, err := st.Save(model, dt, duration, seed, integrator, controller, result)
	if err != nil {
		return err
	}

	fmt.Printf("completed in %v\n", elapsed)
	fmt.Printf("run id: %s\n", runID)
	fmt.Printf("steps: %d\n", len(result.States))
	fmt.Println("\nmetrics:")
	for name, val := range result.Metrics {
		fmt.Printf("  %s: %.6f\n", name, val)
	}

	return nil
}

func makeNBodyInitialState(n int) []float64 {
	state := make([]float64, n*4)

	for i := 0; i < n; i++ {
		angle := float64(i) * 2.0 * 3.14159 / float64(n)
		radius := 1.0

		state[i*4] = radius * cosApprox(angle)
		state[i*4+1] = radius * sinApprox(angle)
		state[i*4+2] = -sinApprox(angle) * 0.5
		state[i*4+3] = cosApprox(angle) * 0.5
	}

	return state
}

func cosApprox(x float64) float64 {
	for x > 3.14159 {
		x -= 2 * 3.14159
	}
	for x < -3.14159 {
		x += 2 * 3.14159
	}
	return 1 - x*x/2 + x*x*x*x/24
}

func sinApprox(x float64) float64 {
	for x > 3.14159 {
		x -= 2 * 3.14159
	}
	for x < -3.14159 {
		x += 2 * 3.14159
	}
	return x - x*x*x/6 + x*x*x*x*x/120
}

func listRuns(cmd *cobra.Command, args []string) error {
	st := storage.New(dataDir)
	runs, err := st.List()
	if err != nil {
		return err
	}

	if len(runs) == 0 {
		fmt.Println("no runs found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tMODEL\tTIME\tDURATION\tDT\tINTEG\tCTRL")

	for _, run := range runs {
		fmt.Fprintf(w, "%s\t%s\t%s\t%.2fs\t%.4fs\t%s\t%s\n",
			run.ID,
			run.Model,
			run.Timestamp.Format("2006-01-02 15:04:05"),
			run.Duration,
			run.Dt,
			run.Integrator,
			run.Controller,
		)
	}

	return w.Flush()
}

func plotRun(cmd *cobra.Command, args []string) error {
	runID := args[0]

	st := storage.New(dataDir)
	meta, err := st.Load(runID)
	if err != nil {
		return err
	}

	states, times, err := st.LoadStates(runID)
	if err != nil {
		return err
	}

	if len(states) == 0 {
		return fmt.Errorf("no data to plot")
	}

	fmt.Printf("run: %s\n", meta.ID)
	fmt.Printf("model: %s\n", meta.Model)
	fmt.Printf("samples: %d\n\n", len(states))

	numVars := len(states[0])
	maxPlots := 6
	if numVars > maxPlots {
		numVars = maxPlots
	}

	for varIdx := 0; varIdx < numVars; varIdx++ {
		data := make([]float64, len(states))
		for i := range states {
			if varIdx < len(states[i]) {
				data[i] = states[i][varIdx]
			}
		}

		caption := fmt.Sprintf("x%d vs time", varIdx)
		if meta.Model == "pendulum" {
			if varIdx == 0 {
				caption = "theta (angle)"
			} else if varIdx == 1 {
				caption = "omega (angular velocity)"
			}
		} else if meta.Model == "cartpole" {
			if varIdx == 0 {
				caption = "cart position"
			} else if varIdx == 1 {
				caption = "cart velocity"
			} else if varIdx == 2 {
				caption = "pole angle"
			} else if varIdx == 3 {
				caption = "pole angular velocity"
			}
		}

		graph := asciigraph.Plot(data,
			asciigraph.Height(10),
			asciigraph.Width(80),
			asciigraph.Caption(caption),
		)
		fmt.Println(graph)
		fmt.Println()
	}

	_ = times

	return nil
}

func exportRun(cmd *cobra.Command, args []string) error {
	runID := args[0]

	st := storage.New(dataDir)
	meta, err := st.Load(runID)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(meta)
}

func benchModel(cmd *cobra.Command, args []string) error {
	model := args[0]

	registry := experiment.NewRegistry()
	dyn, err := registry.GetModel(model)
	if err != nil {
		return err
	}

	integ, err := registry.GetIntegrator("rk4")
	if err != nil {
		return err
	}

	ctrl, err := registry.GetController("none", map[string]float64{"dim": 1})
	if err != nil {
		return err
	}

	durations := []float64{1.0, 5.0, 10.0}
	dts := []float64{0.001, 0.01, 0.1}

	var initState []float64
	switch model {
	case "cartpole":
		initState = []float64{0.0, 0.0, 0.1, 0.0}
	case "nbody":
		initState = makeNBodyInitialState(3)
	default:
		initState = []float64{0.5, 0.0}
	}

	fmt.Printf("benchmarking %s\n\n", model)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "DURATION\tDT\tSTEPS\tTIME\tSTEPS/SEC")

	for _, dur := range durations {
		for _, dt := range dts {
			cfg := experiment.Config{
				Model:      model,
				Integrator: "rk4",
				Controller: "none",
				InitState:  initState,
				Dt:         dt,
				Duration:   dur,
				Seed:       42,
			}

			exp := experiment.New(cfg)
			if err := exp.Setup(dyn, integ, ctrl, nil); err != nil {
				return err
			}

			start := time.Now()
			result, err := exp.Run(context.Background())
			if err != nil {
				return err
			}
			elapsed := time.Since(start)

			steps := len(result.States)
			stepsPerSec := float64(steps) / elapsed.Seconds()

			fmt.Fprintf(w, "%.1fs\t%.4fs\t%d\t%v\t%.0f\n",
				dur, dt, steps, elapsed, stepsPerSec)
		}
	}

	return w.Flush()
}

func analyzeRun(cmd *cobra.Command, args []string) error {
	runID := args[0]

	st := storage.New(dataDir)
	meta, err := st.Load(runID)
	if err != nil {
		return err
	}

	states, _, err := st.LoadStates(runID)
	if err != nil {
		return err
	}

	if len(states) == 0 || len(states[0]) == 0 {
		return fmt.Errorf("no data")
	}

	fmt.Printf("frequency analysis: %s\n", meta.ID)
	fmt.Printf("model: %s\n\n", meta.Model)

	data := make([]float64, len(states))
	for i := range states {
		data[i] = states[i][0]
	}

	n := 1
	for n < len(data) {
		n *= 2
	}
	padded := make([]float64, n)
	copy(padded, data)

	ps := analysis.PowerSpectrum(padded)

	plotData := ps[:len(ps)/4]

	graph := asciigraph.Plot(plotData,
		asciigraph.Height(15),
		asciigraph.Width(80),
		asciigraph.Caption("power spectrum (x0)"),
	)
	fmt.Println(graph)
	fmt.Println()

	maxPower := 0.0
	maxIdx := 0
	for i := 1; i < len(plotData); i++ {
		if plotData[i] > maxPower {
			maxPower = plotData[i]
			maxIdx = i
		}
	}

	freq := float64(maxIdx) / (meta.Duration)
	fmt.Printf("dominant frequency: %.3f hz\n", freq)
	if freq > 0 {
		fmt.Printf("period: %.3f s\n", 1.0/freq)
	}

	return nil
}

func runLive(cmd *cobra.Command, args []string) error {
	model := args[0]

	registry := experiment.NewRegistry()

	dyn, err := registry.GetModel(model)
	if err != nil {
		return err
	}

	integ, err := registry.GetIntegrator(integrator)
	if err != nil {
		return err
	}

	controllerParams := map[string]float64{
		"dim":    float64(dyn.ControlDim()),
		"kp":     kp,
		"ki":     ki,
		"kd":     kd,
		"target": target,
	}
	ctrl, err := registry.GetController(controller, controllerParams)
	if err != nil {
		return err
	}

	var initState []float64
	switch model {
	case "cartpole":
		initState = []float64{pos, vel, theta, omega}
	case "double_pendulum":
		initState = []float64{theta, theta2, omega, omega2}
	case "spring_mass":
		initState = []float64{pos, vel}
	case "spring_chain":
		initState = []float64{pos, 0, 0, vel, 0, 0}
	case "drone":
		initState = []float64{0, 5, theta, 0, 0, omega}
	case "nbody":
		initState = makeNBodyInitialState(numBodies)
	default:
		initState = []float64{theta, omega}
	}

	// Initialize TUI Model
	m := viz.NewModel(dyn, integ, ctrl, initState, dt, model)

	// Run Bubble Tea Program
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

func phasePlot(cmd *cobra.Command, args []string) error {
	runID := args[0]

	st := storage.New(dataDir)
	meta, err := st.Load(runID)
	if err != nil {
		return err
	}

	states, _, err := st.LoadStates(runID)
	if err != nil {
		return err
	}

	if len(states) == 0 {
		return fmt.Errorf("no data to plot")
	}

	if len(states[0]) <= xAxis || len(states[0]) <= yAxis {
		return fmt.Errorf("state dimension too small for selected axes")
	}

	fmt.Printf("phase space plot: %s\n", meta.ID)
	fmt.Printf("model: %s\n", meta.Model)
	fmt.Printf("x-axis: x%d, y-axis: x%d\n\n", xAxis, yAxis)

	// Extract data for phase plot
	xData := make([]float64, len(states))
	yData := make([]float64, len(states))
	for i := range states {
		xData[i] = states[i][xAxis]
		yData[i] = states[i][yAxis]
	}

	// Find bounds
	xMin, xMax := xData[0], xData[0]
	yMin, yMax := yData[0], yData[0]
	for i := range xData {
		if xData[i] < xMin {
			xMin = xData[i]
		}
		if xData[i] > xMax {
			xMax = xData[i]
		}
		if yData[i] < yMin {
			yMin = yData[i]
		}
		if yData[i] > yMax {
			yMax = yData[i]
		}
	}

	// Add padding
	xRange := xMax - xMin
	yRange := yMax - yMin
	if xRange == 0 {
		xRange = 1
	}
	if yRange == 0 {
		yRange = 1
	}

	// Create ASCII scatter plot
	width := 70
	height := 20
	canvas := make([][]rune, height)
	for i := range canvas {
		canvas[i] = make([]rune, width)
		for j := range canvas[i] {
			canvas[i][j] = ' '
		}
	}

	// Plot points
	for i := range xData {
		px := int(float64(width-1) * (xData[i] - xMin) / xRange)
		py := int(float64(height-1) * (yData[i] - yMin) / yRange)
		py = height - 1 - py // Flip y-axis
		if px >= 0 && px < width && py >= 0 && py < height {
			// Use different characters based on density/time
			if i < len(xData)/3 {
				canvas[py][px] = '.'
			} else if i < 2*len(xData)/3 {
				canvas[py][px] = 'o'
			} else {
				canvas[py][px] = '●'
			}
		}
	}

	// Draw frame
	fmt.Printf("  %.2f ┌", yMax)
	for i := 0; i < width; i++ {
		fmt.Print("─")
	}
	fmt.Println("┐")

	for i := range canvas {
		if i == height/2 {
			fmt.Printf("  %.2f │", (yMax+yMin)/2)
		} else {
			fmt.Print("       │")
		}
		fmt.Print(string(canvas[i]))
		fmt.Println("│")
	}

	fmt.Printf("  %.2f └", yMin)
	for i := 0; i < width; i++ {
		fmt.Print("─")
	}
	fmt.Println("┘")

	fmt.Printf("       %.2f", xMin)
	padding := width - 20
	for i := 0; i < padding; i++ {
		fmt.Print(" ")
	}
	fmt.Printf("%.2f\n", xMax)

	fmt.Printf("\nLegend: . = early, o = middle, ● = late\n")

	return nil
}

func exportCSV(cmd *cobra.Command, args []string) error {
	runID := args[0]

	st := storage.New(dataDir)
	states, times, err := st.LoadStates(runID)
	if err != nil {
		return err
	}

	if len(states) == 0 {
		return fmt.Errorf("no data to export")
	}

	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	// Header
	header := []string{"time"}
	for i := range states[0] {
		header = append(header, fmt.Sprintf("x%d", i))
	}
	if err := w.Write(header); err != nil {
		return err
	}

	// Data rows
	for i := range states {
		row := []string{strconv.FormatFloat(times[i], 'f', 6, 64)}
		for _, val := range states[i] {
			row = append(row, strconv.FormatFloat(val, 'f', 6, 64))
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}

	return nil
}

func compareIntegrators(cmd *cobra.Command, args []string) error {
	model := args[0]
	integrators := args[1:]

	registry := experiment.NewRegistry()
	dyn, err := registry.GetModel(model)
	if err != nil {
		return err
	}

	initState := []float64{theta, 0}
	switch model {
	case "double_pendulum":
		initState = []float64{theta, theta, 0, 0}
	case "cartpole":
		initState = []float64{0, 0, theta, 0}
	case "nbody":
		n := 3
		initState = make([]float64, n*4)
		for i := 0; i < n; i++ {
			angle := float64(i) * 2.0 * 3.14159 / float64(n)
			initState[i*4] = 2.0 * float64(i+1) * 0.5
			initState[i*4+1] = 0
			initState[i*4+2] = 0
			initState[i*4+3] = 0.5 * float64(i+1) * 0.3 * angle
		}
	case "drone":
		initState = []float64{0, 5, theta, 0, 0, 0}
	case "spring_mass":
		initState = []float64{1.0, 0}
	}

	fmt.Printf("comparing integrators for %s (dt=%.4f, duration=%.1fs)\n\n", model, dt, duration)
	fmt.Printf("%-12s  %-12s  %-12s  %-12s\n", "integrator", "final_x0", "energy_drift", "time_ms")
	fmt.Println(strings.Repeat("-", 52))

	for _, intName := range integrators {
		integ, err := registry.GetIntegrator(intName)
		if err != nil {
			fmt.Printf("%-12s  error: %v\n", intName, err)
			continue
		}

		ctrl := control.NewNone(dyn.ControlDim())
		s := dynamo.New(dyn, integ, ctrl)

		cfg := dynamo.Config{Dt: dt, Duration: duration}

		start := time.Now()
		result, err := s.Run(context.Background(), initState, cfg)
		elapsed := time.Since(start)

		if err != nil {
			fmt.Printf("%-12s  error: %v\n", intName, err)
			continue
		}

		finalX0 := 0.0
		if len(result.States) > 0 && len(result.States[len(result.States)-1]) > 0 {
			finalX0 = result.States[len(result.States)-1][0]
		}

		fmt.Printf("%-12s  %12.6f  %12.2e  %12.2f\n", intName, finalX0, result.EnergyDrift, float64(elapsed.Microseconds())/1000)
	}

	return nil
}

func exportJSON(cmd *cobra.Command, args []string) error {
	runID := args[0]

	st := storage.New(dataDir)
	meta, err := st.Load(runID)
	if err != nil {
		return err
	}

	states, times, err := st.LoadStates(runID)
	if err != nil {
		return err
	}

	result := &dynamo.Result{
		States:  make([]dynamo.State, len(states)),
		Times:   times,
		Metrics: meta.Metrics,
	}
	for i, s := range states {
		result.States[i] = s
	}

	return storage.ExportJSONStdout(meta.ID, meta.Model, meta.Integrator, meta.Controller, meta.Dt, meta.Duration, result)
}