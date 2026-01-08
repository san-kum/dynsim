package analysis

import (
	"math"

	"github.com/san-kum/dynsim/internal/dynamo"
)

// LyapunovExponent estimates the largest Lyapunov exponent using the
// trajectory separation method. A positive value indicates chaos.
//
// Algorithm:
// 1. Run two nearby trajectories
// 2. Measure their divergence over time
// 3. λ ≈ (1/t) * ln(|δx(t)/δx(0)|)
func LyapunovExponent(
	dyn dynamo.System,
	integ dynamo.Integrator,
	x0 dynamo.State,
	dt, duration float64,
	perturbation float64,
) float64 {
	if len(x0) == 0 {
		return 0
	}

	// Create perturbed initial condition
	x0p := make(dynamo.State, len(x0))
	copy(x0p, x0)
	x0p[0] += perturbation

	// Initial separation
	d0 := perturbation

	// Run simulation
	x := make(dynamo.State, len(x0))
	xp := make(dynamo.State, len(x0))
	copy(x, x0)
	copy(xp, x0p)

	ctrl := make(dynamo.Control, dyn.ControlDim())
	t := 0.0

	sumLog := 0.0
	count := 0

	for t < duration {
		x = integ.Step(dyn, x, ctrl, t, dt)
		xp = integ.Step(dyn, xp, ctrl, t, dt)
		t += dt

		// Calculate separation
		sep := 0.0
		for i := range x {
			diff := xp[i] - x[i]
			sep += diff * diff
		}
		sep = math.Sqrt(sep)

		if sep > 0 && d0 > 0 {
			sumLog += math.Log(sep / d0)
			count++
		}

		// Renormalize to prevent overflow
		if sep > 1.0 {
			scale := d0 / sep
			for i := range xp {
				xp[i] = x[i] + (xp[i]-x[i])*scale
			}
		}
	}

	if count == 0 || t == 0 {
		return 0
	}

	return sumLog / (float64(count) * dt)
}

// LyapunovSpectrum computes multiple Lyapunov exponents by perturbing
// each state dimension independently.
func LyapunovSpectrum(
	dyn dynamo.System,
	integ dynamo.Integrator,
	x0 dynamo.State,
	dt, duration float64,
	perturbation float64,
) []float64 {
	n := len(x0)
	spectrum := make([]float64, n)

	for i := 0; i < n; i++ {
		xp := make(dynamo.State, n)
		copy(xp, x0)
		xp[i] += perturbation

		spectrum[i] = lyapunovForPerturbation(dyn, integ, x0, xp, dt, duration, perturbation)
	}

	return spectrum
}

func lyapunovForPerturbation(
	dyn dynamo.System,
	integ dynamo.Integrator,
	x0, x0p dynamo.State,
	dt, duration, d0 float64,
) float64 {
	x := make(dynamo.State, len(x0))
	xp := make(dynamo.State, len(x0p))
	copy(x, x0)
	copy(xp, x0p)

	ctrl := make(dynamo.Control, dyn.ControlDim())
	t := 0.0

	sumLog := 0.0
	count := 0

	for t < duration {
		x = integ.Step(dyn, x, ctrl, t, dt)
		xp = integ.Step(dyn, xp, ctrl, t, dt)
		t += dt

		sep := 0.0
		for i := range x {
			diff := xp[i] - x[i]
			sep += diff * diff
		}
		sep = math.Sqrt(sep)

		if sep > 0 && d0 > 0 {
			sumLog += math.Log(sep / d0)
			count++
		}

		if sep > 1.0 {
			scale := d0 / sep
			for i := range xp {
				xp[i] = x[i] + (xp[i]-x[i])*scale
			}
		}
	}

	if count == 0 {
		return 0
	}
	return sumLog / (float64(count) * dt)
}
