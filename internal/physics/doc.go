// Package physics provides dynamical system models for simulation.
//
// Each model implements the [dynamo.System] interface, defining the
// differential equations governing the system's evolution:
//
//   - [Pendulum]: simple harmonic oscillator
//   - [DoublePendulum]: chaotic coupled pendulum
//   - [Lorenz]: butterfly attractor
//   - [ThreeBody]: gravitational three-body problem
//   - [NBody]: N-particle gravitational simulation (GPU-accelerated)
//
// Many models also implement [dynamo.Configurable] for runtime parameter
// adjustment and [dynamo.Hamiltonian] for energy calculation.
//
// # Energy Conservation
//
// For Hamiltonian systems, use [dynamo.Hamiltonian] to monitor energy drift:
//
//	dyn := physics.NewPendulum()
//	if h, ok := dyn.(dynamo.Hamiltonian); ok {
//	    energy := h.Energy(state)
//	}
package physics
