//go:build !cuda

package compute

type CUDABackend struct{}

func NewCUDABackend() *CUDABackend {
	return &CUDABackend{}
}

func (c *CUDABackend) Name() string    { return "cuda (not available)" }
func (c *CUDABackend) Available() bool { return false }
func (c *CUDABackend) Cleanup()        {}

func (c *CUDABackend) NBodyForces(positions []float64, masses []float64, g, softening float64) ([]float64, []float64) {
	cpu := NewCPUBackend()
	return cpu.NBodyForces(positions, masses, g, softening)
}

func (c *CUDABackend) MatVecMul(mat [][]float64, vec []float64) []float64 {
	cpu := NewCPUBackend()
	return cpu.MatVecMul(mat, vec)
}
