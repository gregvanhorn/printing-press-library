.PHONY: build test lint install clean

build:
	go build -o bin/trigger-dev-pp-cli ./cmd/trigger-dev-pp-cli

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/trigger-dev-pp-cli

clean:
	rm -rf bin/

build-mcp:
	go build -o bin/trigger-dev-pp-mcp ./cmd/trigger-dev-pp-mcp

install-mcp:
	go install ./cmd/trigger-dev-pp-mcp

build-all: build build-mcp
