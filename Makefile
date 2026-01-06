.PHONY: build test clean run bench

build:
	go build -o bin/dynsim cmd/dynsim/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/ .dynsim/

run:
	go run cmd/dynsim/main.go run pendulum

bench:
	go run cmd/dynsim/main.go bench pendulum

install:
	go install cmd/dynsim/main.go
