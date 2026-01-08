#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPUTE_DIR="$SCRIPT_DIR/internal/compute"

echo "building cuda kernels..."

cd "$COMPUTE_DIR"

nvcc -c -o kernels.o kernels.cu \
    -arch=sm_75 \
    -O3 \
    --compiler-options '-fPIC'

ar rcs libkernels.a kernels.o

rm kernels.o

echo "libkernels.a created"

cd "$SCRIPT_DIR"

echo "building dynsim with cuda..."

CGO_LDFLAGS="-L$COMPUTE_DIR" go build -tags cuda -o dynsim cmd/dynsim/main.go

echo "done. run ./dynsim"
