package compute

type Backend interface {
	Name() string
	Available() bool
	NBodyForces(positions []float64, masses []float64, g, softening float64) (ax, ay []float64)
	MatVecMul(mat [][]float64, vec []float64) []float64
	Cleanup()
}

var activeBackend Backend

func init() {
	// Auto-select best available backend (CUDA if available, else CPU)
	activeBackend = AutoSelectBackend()
}

func SetBackend(b Backend) {
	if activeBackend != nil {
		activeBackend.Cleanup()
	}
	activeBackend = b
}

func GetBackend() Backend {
	return activeBackend
}

func AutoSelectBackend() Backend {
	cuda := NewCUDABackend()
	if cuda.Available() {
		return cuda
	}
	return NewCPUBackend()
}
