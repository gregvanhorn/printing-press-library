.PHONY: build test lint install clean

build:
	go build -o bin/espn-pp-cli ./cmd/espn-pp-cli

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/espn-pp-cli

clean:
	rm -rf bin/

build-mcp:
	go build -o bin/espn-pp-mcp ./cmd/espn-pp-mcp

install-mcp:
	go install ./cmd/espn-pp-mcp

build-all: build build-mcp
