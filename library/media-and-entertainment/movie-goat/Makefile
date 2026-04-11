.PHONY: build test lint install clean

build:
	go build -o bin/movie-goat-pp-cli ./cmd/movie-goat-pp-cli

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/movie-goat-pp-cli

clean:
	rm -rf bin/

build-mcp:
	go build -o bin/movie-goat-pp-mcp ./cmd/movie-goat-pp-mcp

install-mcp:
	go install ./cmd/movie-goat-pp-mcp

build-all: build build-mcp
