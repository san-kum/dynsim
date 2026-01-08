package integrators

import (
	"math"

	"github.com/san-kum/dynsim/internal/dynamo"
)

// Dormand-Prince coefficients (RK45)
var (
	a2 = 1.0 / 5.0
	a3 = 3.0 / 10.0
	a4 = 4.0 / 5.0
	a5 = 8.0 / 9.0

	b21 = 1.0 / 5.0
	b31 = 3.0 / 40.0
	b32 = 9.0 / 40.0
	b41 = 44.0 / 45.0
	b42 = -56.0 / 15.0
	b43 = 32.0 / 9.0
	b51 = 19372.0 / 6561.0
	b52 = -25360.0 / 2187.0
	b53 = 64448.0 / 6561.0
	b54 = -212.0 / 729.0
	b61 = 9017.0 / 3168.0
	b62 = -355.0 / 33.0
	b63 = 46732.0 / 5247.0
	b64 = 49.0 / 176.0
	b65 = -5103.0 / 18656.0

	c1 = 35.0 / 384.0
	c3 = 500.0 / 1113.0
	c4 = 125.0 / 192.0
	c5 = -2187.0 / 6784.0
	c6 = 11.0 / 84.0

	dc1 = c1 - 5179.0/57600.0
	dc3 = c3 - 7571.0/16695.0
	dc4 = c4 - 393.0/640.0
	dc5 = c5 - -92097.0/339200.0
	dc6 = c6 - 187.0/2100.0
	dc7 = -1.0 / 40.0
)

type RK45 struct {
	safety   float64
	minScale float64
	maxScale float64
}

func NewRK45() *RK45 {
	return &RK45{
		safety:   0.9,
		minScale: 0.2,
		maxScale: 10.0,
	}
}

func (r *RK45) Step(dyn dynamo.System, x dynamo.State, u dynamo.Control, t, dt float64) dynamo.State {
	newX, _, _ := r.StepAdaptive(dyn, x, u, t, dt, 1e-6)
	return newX
}

func (r *RK45) StepAdaptive(dyn dynamo.System, x dynamo.State, u dynamo.Control, t, dt, tol float64) (dynamo.State, float64, error) {
	n := len(x)

	k1 := dyn.Derive(x, u, t)

	x2 := make(dynamo.State, n)
	for i := 0; i < n; i++ {
		x2[i] = x[i] + dt*b21*k1[i]
	}
	k2 := dyn.Derive(x2, u, t+a2*dt)

	x3 := make(dynamo.State, n)
	for i := 0; i < n; i++ {
		x3[i] = x[i] + dt*(b31*k1[i]+b32*k2[i])
	}
	k3 := dyn.Derive(x3, u, t+a3*dt)

	x4 := make(dynamo.State, n)
	for i := 0; i < n; i++ {
		x4[i] = x[i] + dt*(b41*k1[i]+b42*k2[i]+b43*k3[i])
	}
	k4 := dyn.Derive(x4, u, t+a4*dt)

	x5 := make(dynamo.State, n)
	for i := 0; i < n; i++ {
		x5[i] = x[i] + dt*(b51*k1[i]+b52*k2[i]+b53*k3[i]+b54*k4[i])
	}
	k5 := dyn.Derive(x5, u, t+a5*dt)

	x6 := make(dynamo.State, n)
	for i := 0; i < n; i++ {
		x6[i] = x[i] + dt*(b61*k1[i]+b62*k2[i]+b63*k3[i]+b64*k4[i]+b65*k5[i])
	}
	k6 := dyn.Derive(x6, u, t+dt)

	xNew := make(dynamo.State, n)
	for i := 0; i < n; i++ {
		xNew[i] = x[i] + dt*(c1*k1[i]+c3*k3[i]+c4*k4[i]+c5*k5[i]+c6*k6[i])
	}

	k7 := dyn.Derive(xNew, u, t+dt)

	errMax := 0.0
	for i := 0; i < n; i++ {
		errEst := dt * (dc1*k1[i] + dc3*k3[i] + dc4*k4[i] + dc5*k5[i] + dc6*k6[i] + dc7*k7[i])
		scale := math.Abs(x[i]) + math.Abs(dt*k1[i]) + 1e-10
		errMax = math.Max(errMax, math.Abs(errEst)/scale)
	}

	errRatio := errMax / tol

	var dtNew float64
	if errRatio > 1 {
		scale := math.Max(r.minScale, r.safety*math.Pow(errRatio, -0.25))
		dtNew = dt * scale
	} else {
		if errRatio > 0 {
			scale := math.Min(r.maxScale, r.safety*math.Pow(errRatio, -0.2))
			dtNew = dt * scale
		} else {
			dtNew = dt * r.maxScale
		}
	}

	return xNew, dtNew, nil
}
