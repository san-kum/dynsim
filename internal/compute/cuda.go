//go:build cuda

package compute

/*
#cgo CFLAGS: -I/opt/cuda/include
#cgo LDFLAGS: -L/opt/cuda/lib64 -L${SRCDIR} -lcudart -lkernels -lstdc++
#include <stdlib.h>

extern int cuda_device_count();
extern const char* cuda_device_name_get();
extern void nbody_gpu(float* positions, float* masses, float* ax, float* ay, int n, float g, float softening);
*/
import "C"
import "unsafe"

type CUDABackend struct {
	available  bool
	deviceName string
}

func NewCUDABackend() *CUDABackend {
	count := int(C.cuda_device_count())
	name := ""
	if count > 0 {
		name = C.GoString(C.cuda_device_name_get())
	}
	return &CUDABackend{
		available:  count > 0,
		deviceName: name,
	}
}

func (c *CUDABackend) Name() string {
	if c.available {
		return "cuda (" + c.deviceName + ")"
	}
	return "cuda (not available)"
}

func (c *CUDABackend) Available() bool { return c.available }
func (c *CUDABackend) Cleanup()        {}

func (c *CUDABackend) NBodyForces(positions []float64, masses []float64, g, softening float64) ([]float64, []float64) {
	if !c.available {
		cpu := NewCPUBackend()
		return cpu.NBodyForces(positions, masses, g, softening)
	}

	n := len(masses)
	ax := make([]float64, n)
	ay := make([]float64, n)

	posF := make([]float32, len(positions))
	massF := make([]float32, n)
	axF := make([]float32, n)
	ayF := make([]float32, n)

	for i := range positions {
		posF[i] = float32(positions[i])
	}
	for i := range masses {
		massF[i] = float32(masses[i])
	}

	C.nbody_gpu(
		(*C.float)(unsafe.Pointer(&posF[0])),
		(*C.float)(unsafe.Pointer(&massF[0])),
		(*C.float)(unsafe.Pointer(&axF[0])),
		(*C.float)(unsafe.Pointer(&ayF[0])),
		C.int(n),
		C.float(g),
		C.float(softening),
	)

	for i := 0; i < n; i++ {
		ax[i] = float64(axF[i])
		ay[i] = float64(ayF[i])
	}

	return ax, ay
}

func (c *CUDABackend) MatVecMul(mat [][]float64, vec []float64) []float64 {
	cpu := NewCPUBackend()
	return cpu.MatVecMul(mat, vec)
}
