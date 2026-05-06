.PHONY: build test lint install clean

build:
	go build -o bin/amazon-sp-api-pp-cli ./cmd/amazon-sp-api-pp-cli

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/amazon-sp-api-pp-cli

clean:
	rm -rf bin/

build-mcp:
	go build -o bin/amazon-sp-api-pp-mcp ./cmd/amazon-sp-api-pp-mcp

install-mcp:
	go install ./cmd/amazon-sp-api-pp-mcp

build-all: build build-mcp
