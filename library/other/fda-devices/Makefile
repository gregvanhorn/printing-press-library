.PHONY: build test lint install clean

build:
	go build -o bin/fda-devices-pp-cli ./cmd/fda-devices-pp-cli

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/fda-devices-pp-cli

clean:
	rm -rf bin/

build-mcp:
	go build -o bin/fda-devices-pp-mcp ./cmd/fda-devices-pp-mcp

install-mcp:
	go install ./cmd/fda-devices-pp-mcp

build-all: build build-mcp
