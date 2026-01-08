// Package analysis provides chaos and dynamics analysis tools.
//
// The package includes tools for characterizing dynamical systems:
//
//   - [LyapunovExponent]: largest Lyapunov exponent via trajectory separation
//   - [LyapunovSpectrum]: full Lyapunov spectrum for multi-dimensional systems
//   - [BifurcationDiagram]: parameter sweep for bifurcation analysis
//   - [GeneratePhasePortrait]: 2D phase space trajectories
//   - [GeneratePoincareSection]: stroboscopic section of phase space
//
// # Chaos Detection
//
// A positive largest Lyapunov exponent indicates chaotic dynamics:
//
//	lambda := analysis.LyapunovExponent(dyn, integ, x0, dt, duration)
//	if lambda > 0 {
//	    // System is chaotic
//	}
package analysis
