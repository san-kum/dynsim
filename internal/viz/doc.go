// Package viz provides terminal-based visualization for dynamical simulations.
//
// The package implements an interactive TUI using the Bubble Tea framework:
//
//   - [App]: main interactive application with model selection
//   - [Canvas]: Braille-based pixel canvas for high-fidelity rendering
//   - Theme selection with 5 built-in color schemes
//
// # Key Bindings
//
//	Space - Pause/Resume simulation
//	R     - Reset to initial state
//	T     - Cycle color themes
//	G     - Toggle GIF recording
//	?     - Show help overlay
//	[]/   - Time travel (rewind/forward)
//
// # Recording
//
// The visualization supports recording simulation sessions as GIF animations
// using the G key. Recordings are saved to the current directory.
package viz
