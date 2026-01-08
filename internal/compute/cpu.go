package compute

import (
	"math"
	"runtime"
	"sync"
)

type CPUBackend struct {
	workers int
}

func NewCPUBackend() *CPUBackend {
	return &CPUBackend{
		workers: runtime.NumCPU(),
	}
}

func (c *CPUBackend) Name() string    { return "cpu" }
func (c *CPUBackend) Available() bool { return true }
func (c *CPUBackend) Cleanup()        {}

func (c *CPUBackend) NBodyForces(positions []float64, masses []float64, g, softening float64) ([]float64, []float64) {
	n := len(masses)
	ax := make([]float64, n)
	ay := make([]float64, n)

	if n < 16 {
		c.nbodySerial(positions, masses, g, softening, ax, ay)
		return ax, ay
	}

	c.nbodyParallel(positions, masses, g, softening, ax, ay)
	return ax, ay
}

func (c *CPUBackend) nbodySerial(pos []float64, masses []float64, g, eps float64, ax, ay []float64) {
	n := len(masses)
	eps2 := eps * eps

	for i := 0; i < n; i++ {
		xi, yi := pos[i*2], pos[i*2+1]

		for j := i + 1; j < n; j++ {
			xj, yj := pos[j*2], pos[j*2+1]

			rx := xj - xi
			ry := yj - yi
			r2 := rx*rx + ry*ry + eps2

			rInv := 1.0 / math.Sqrt(r2)
			r3Inv := rInv * rInv * rInv

			fij := g * masses[j] * r3Inv
			ax[i] += fij * rx
			ay[i] += fij * ry

			fji := g * masses[i] * r3Inv
			ax[j] -= fji * rx
			ay[j] -= fji * ry
		}
	}
}

func (c *CPUBackend) nbodyParallel(pos []float64, masses []float64, g, eps float64, ax, ay []float64) {
	n := len(masses)
	eps2 := eps * eps

	localAx := make([][]float64, c.workers)
	localAy := make([][]float64, c.workers)
	for w := 0; w < c.workers; w++ {
		localAx[w] = make([]float64, n)
		localAy[w] = make([]float64, n)
	}

	var wg sync.WaitGroup
	chunkSize := (n + c.workers - 1) / c.workers

	for w := 0; w < c.workers; w++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()

			start := worker * chunkSize
			end := start + chunkSize
			if end > n {
				end = n
			}

			lax := localAx[worker]
			lay := localAy[worker]

			for i := start; i < end; i++ {
				xi, yi := pos[i*2], pos[i*2+1]

				for j := 0; j < n; j++ {
					if i == j {
						continue
					}

					xj, yj := pos[j*2], pos[j*2+1]

					rx := xj - xi
					ry := yj - yi
					r2 := rx*rx + ry*ry + eps2

					rInv := 1.0 / math.Sqrt(r2)
					r3Inv := rInv * rInv * rInv

					f := g * masses[j] * r3Inv
					lax[i] += f * rx
					lay[i] += f * ry
				}
			}
		}(w)
	}

	wg.Wait()

	for w := 0; w < c.workers; w++ {
		for i := 0; i < n; i++ {
			ax[i] += localAx[w][i]
			ay[i] += localAy[w][i]
		}
	}
}

func (c *CPUBackend) MatVecMul(mat [][]float64, vec []float64) []float64 {
	rows := len(mat)
	result := make([]float64, rows)

	if rows < 16 {
		for i := 0; i < rows; i++ {
			sum := 0.0
			for j := 0; j < len(vec); j++ {
				if j < len(mat[i]) {
					sum += mat[i][j] * vec[j]
				}
			}
			result[i] = sum
		}
		return result
	}

	var wg sync.WaitGroup
	chunkSize := (rows + c.workers - 1) / c.workers

	for w := 0; w < c.workers; w++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()

			start := worker * chunkSize
			end := start + chunkSize
			if end > rows {
				end = rows
			}

			for i := start; i < end; i++ {
				sum := 0.0
				for j := 0; j < len(vec); j++ {
					if j < len(mat[i]) {
						sum += mat[i][j] * vec[j]
					}
				}
				result[i] = sum
			}
		}(w)
	}

	wg.Wait()
	return result
}
