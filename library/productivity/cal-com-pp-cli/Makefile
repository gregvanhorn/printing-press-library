.PHONY: build test lint install clean

build:
	go build -o bin/cal-com-pp-cli ./cmd/cal-com-pp-cli

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/cal-com-pp-cli

clean:
	rm -rf bin/

build-mcp:
	go build -o bin/cal-com-mcp ./cmd/cal-com-mcp

install-mcp:
	go install ./cmd/cal-com-mcp

build-all: build build-mcp
