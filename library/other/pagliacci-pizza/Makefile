.PHONY: build test lint install clean

build:
	go build -o bin/pagliacci-pizza-pp-cli ./cmd/pagliacci-pizza-pp-cli

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/pagliacci-pizza-pp-cli

clean:
	rm -rf bin/

build-mcp:
	go build -o bin/pagliacci-pizza-mcp ./cmd/pagliacci-pizza-mcp

install-mcp:
	go install ./cmd/pagliacci-pizza-mcp

build-all: build build-mcp
