#include <cuda_runtime.h>
#include <math.h>

#define TILE_SIZE 256

extern "C" {

// Optimized N-Body kernel with shared memory tiling
// Uses tile-based approach for better cache utilization on 1000+ particles
__global__ void nbody_kernel_tiled(
    const float* __restrict__ positions,
    const float* __restrict__ masses,
    float* __restrict__ ax,
    float* __restrict__ ay,
    int n,
    float g,
    float eps2
) {
    // Shared memory for tile of particles
    __shared__ float s_pos[TILE_SIZE * 2];
    __shared__ float s_mass[TILE_SIZE];
    
    int i = blockIdx.x * blockDim.x + threadIdx.x;
    
    float xi = 0.0f, yi = 0.0f;
    if (i < n) {
        xi = positions[i * 2];
        yi = positions[i * 2 + 1];
    }
    
    float axi = 0.0f;
    float ayi = 0.0f;
    
    // Process particles in tiles
    int numTiles = (n + TILE_SIZE - 1) / TILE_SIZE;
    
    for (int tile = 0; tile < numTiles; tile++) {
        // Load tile into shared memory
        int j = tile * TILE_SIZE + threadIdx.x;
        if (j < n) {
            s_pos[threadIdx.x * 2] = positions[j * 2];
            s_pos[threadIdx.x * 2 + 1] = positions[j * 2 + 1];
            s_mass[threadIdx.x] = masses[j];
        } else {
            s_pos[threadIdx.x * 2] = 0.0f;
            s_pos[threadIdx.x * 2 + 1] = 0.0f;
            s_mass[threadIdx.x] = 0.0f;
        }
        __syncthreads();
        
        // Compute forces from this tile
        if (i < n) {
            #pragma unroll 8
            for (int k = 0; k < TILE_SIZE && tile * TILE_SIZE + k < n; k++) {
                int jGlobal = tile * TILE_SIZE + k;
                if (i == jGlobal) continue;
                
                float xj = s_pos[k * 2];
                float yj = s_pos[k * 2 + 1];
                
                float rx = xj - xi;
                float ry = yj - yi;
                float r2 = rx * rx + ry * ry + eps2;
                
                float rInv = rsqrtf(r2);
                float r3Inv = rInv * rInv * rInv;
                float f = g * s_mass[k] * r3Inv;
                
                axi += f * rx;
                ayi += f * ry;
            }
        }
        __syncthreads();
    }
    
    if (i < n) {
        ax[i] = axi;
        ay[i] = ayi;
    }
}

// Simple kernel for small N (< TILE_SIZE)
__global__ void nbody_kernel_simple(
    const float* __restrict__ positions,
    const float* __restrict__ masses,
    float* __restrict__ ax,
    float* __restrict__ ay,
    int n,
    float g,
    float eps2
) {
    int i = blockIdx.x * blockDim.x + threadIdx.x;
    if (i >= n) return;

    float xi = positions[i * 2];
    float yi = positions[i * 2 + 1];
    float axi = 0.0f;
    float ayi = 0.0f;

    for (int j = 0; j < n; j++) {
        if (i == j) continue;

        float xj = positions[j * 2];
        float yj = positions[j * 2 + 1];

        float rx = xj - xi;
        float ry = yj - yi;
        float r2 = rx * rx + ry * ry + eps2;

        float rInv = rsqrtf(r2);
        float r3Inv = rInv * rInv * rInv;
        float f = g * masses[j] * r3Inv;

        axi += f * rx;
        ayi += f * ry;
    }

    ax[i] = axi;
    ay[i] = ayi;
}

void nbody_gpu(
    float* h_positions,
    float* h_masses,
    float* h_ax,
    float* h_ay,
    int n,
    float g,
    float softening
) {
    float *d_positions, *d_masses, *d_ax, *d_ay;
    size_t pos_size = n * 2 * sizeof(float);
    size_t n_size = n * sizeof(float);
    float eps2 = softening * softening;

    cudaMalloc(&d_positions, pos_size);
    cudaMalloc(&d_masses, n_size);
    cudaMalloc(&d_ax, n_size);
    cudaMalloc(&d_ay, n_size);

    cudaMemcpy(d_positions, h_positions, pos_size, cudaMemcpyHostToDevice);
    cudaMemcpy(d_masses, h_masses, n_size, cudaMemcpyHostToDevice);

    // Choose kernel based on problem size
    if (n >= TILE_SIZE) {
        // Use tiled kernel for large N
        int blockSize = TILE_SIZE;
        int numBlocks = (n + blockSize - 1) / blockSize;
        nbody_kernel_tiled<<<numBlocks, blockSize>>>(
            d_positions, d_masses, d_ax, d_ay, n, g, eps2
        );
    } else {
        // Use simple kernel for small N
        int blockSize = 128;
        int numBlocks = (n + blockSize - 1) / blockSize;
        nbody_kernel_simple<<<numBlocks, blockSize>>>(
            d_positions, d_masses, d_ax, d_ay, n, g, eps2
        );
    }

    cudaDeviceSynchronize();

    cudaMemcpy(h_ax, d_ax, n_size, cudaMemcpyDeviceToHost);
    cudaMemcpy(h_ay, d_ay, n_size, cudaMemcpyDeviceToHost);

    cudaFree(d_positions);
    cudaFree(d_masses);
    cudaFree(d_ax);
    cudaFree(d_ay);
}

int cuda_device_count() {
    int count = 0;
    cudaGetDeviceCount(&count);
    return count;
}

const char* cuda_device_name_get() {
    static char name[256] = "unknown";
    cudaDeviceProp prop;
    if (cudaGetDeviceProperties(&prop, 0) == cudaSuccess) {
        strncpy(name, prop.name, 255);
    }
    return name;
}

}
