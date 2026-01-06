package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/guptarohit/asciigraph"
	"github.com/san-kum/dynsim/internal/analysis"
	"github.com/san-kum/dynsim/internal/experiment"
	"github.com/san-kum/dynsim/internal/store"
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
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "dynsim",
		Short: "physics and control simulation lab",
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

	rootCmd.AddCommand(runCmd, listCmd, plotCmd, exportCmd, benchCmd, analyzeCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runSimulation(cmd *cobra.Command, args []string) error {
	model := args[0]

	st := store.New(dataDir)
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
	st := store.New(dataDir)
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

	st := store.New(dataDir)
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

	st := store.New(dataDir)
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

	st := store.New(dataDir)
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
