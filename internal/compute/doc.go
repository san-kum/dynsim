// Package compute provides hardware-accelerated computation backends.
//
// The package automatically selects the best available backend:
//
//   - CUDA: GPU-accelerated N-body force calculation
//   - CPU: Fallback for systems without GPU
//
// # GPU Acceleration
//
// N-body simulations with 32+ particles automatically use GPU when available:
//
//	backend := compute.GetBackend()
//	ax, ay := backend.NBodyForces(positions, masses, g, softening)
//
// Build with CUDA support:
//
//	./build_cuda.sh
//
// For 1000+ particles, GPU provides ~30x speedup over CPU.
package compute
