package dynamo

import "errors"

// Domain errors for simulation operations.
var (
	// ErrInvalidState indicates a state vector with invalid dimensions or values.
	ErrInvalidState = errors.New("dynamo: invalid state (NaN or Inf detected)")

	// ErrUnstable indicates the simulation became numerically unstable.
	ErrUnstable = errors.New("dynamo: simulation unstable (state diverged)")

	// ErrParameterBounds indicates a parameter value is outside valid range.
	ErrParameterBounds = errors.New("dynamo: parameter out of valid bounds")

	// ErrContextCanceled indicates the simulation was interrupted.
	ErrContextCanceled = errors.New("dynamo: simulation canceled by context")

	// ErrStepTooSmall indicates adaptive timestep became too small.
	ErrStepTooSmall = errors.New("dynamo: adaptive timestep below minimum")

	// ErrDimensionMismatch indicates mismatched state/control dimensions.
	ErrDimensionMismatch = errors.New("dynamo: dimension mismatch between state and system")
)

// SimulationError wraps an error with simulation context.
type SimulationError struct {
	Step    int
	Time    float64
	State   State
	Wrapped error
}

func (e *SimulationError) Error() string {
	return e.Wrapped.Error()
}

func (e *SimulationError) Unwrap() error {
	return e.Wrapped
}
