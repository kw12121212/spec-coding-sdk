.PHONY: build test lint fmt clean

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"
BUILD_DIR := ./bin

build:
	go build $(LDFLAGS) -o $(BUILD_DIR)/ ./cmd/...

test:
	go test ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

clean:
	rm -rf $(BUILD_DIR)
