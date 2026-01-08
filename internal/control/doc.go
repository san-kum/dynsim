// Package control provides feedback controllers for dynamical systems.
//
// Controllers implement the [dynamo.Controller] interface to compute
// control inputs based on system state:
//
//   - [PID]: Proportional-Integral-Derivative controller
//   - [LQR]: Linear Quadratic Regulator (requires linearized system)
//   - [None]: Passthrough controller (zero control)
//
// # Usage
//
//	pid := control.NewPID(1.0, 0.1, 0.01, 0.0)  // Kp, Ki, Kd, setpoint
//	sim := dynamo.New(dyn, integ, pid)
//	// Controller.Compute is called each timestep
//
// Controllers implementing [dynamo.Configurable] support live tuning.
package control
