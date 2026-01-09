.PHONY: build test test-all test-integration test-fuzz test-stress clean cuda

build:
	go build -tags cuda -o dynsim cmd/dynsim/main.go

cuda:
	./build_cuda.sh

test:
	go test -v ./tests/...

test-verbose:
	go test -v ./tests/... -ginkgo.v

test-all: test

clean:
	rm -f dynsim libkernels.a internal/compute/libkernels.a internal/compute/kernels.o
