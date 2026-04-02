.PHONY: build test lint install clean

build:
	go build -o bin/steam-web-pp-cli ./cmd/steam-web-pp-cli

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/steam-web-pp-cli

clean:
	rm -rf bin/

build-mcp:
	go build -o bin/steam-web-mcp ./cmd/steam-web-mcp

install-mcp:
	go install ./cmd/steam-web-mcp

build-all: build build-mcp
