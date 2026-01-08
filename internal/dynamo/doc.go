// Package dynamo provides core simulation primitives for dynamical systems.
//
// The package defines the fundamental interfaces and types for numerical
// simulation of ordinary differential equations (ODEs):
//
//   - [State]: vector representing system state
//   - [System]: interface for ODE systems (dX/dt = f(X, u, t))
//   - [Stepper]: numerical integrator interface
//   - [Controller]: feedback controller interface
//   - [Simulator]: orchestrates simulation runs
//
// # Example
//
//	dyn := physics.NewPendulum()
//	integ := integrators.NewRK4()
//	sim := dynamo.New(dyn, integ, nil)
//	result, _ := sim.Run(ctx, x0, cfg)
//
// # Thread Safety
//
// Simulator instances are NOT thread-safe. For parallel simulations,
// use the [Ensemble] type which safely manages multiple simulation runs.
package dynamo
