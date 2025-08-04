# Makefile for Kubernetes Controller Project

# Define variables
REPO_ROOT := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
BIN_DIR := $(REPO_ROOT)/bin
GO := go
GOFMT := gofmt
GOLINT := golangci-lint
MODULE := github.com/PulokSaha0706/my-controller

# Build flags
BUILD_FLAGS := -v

# Kubernetes related variables
KUBECONFIG ?= $(HOME)/.kube/config

.PHONY: all build clean test lint fmt gen verify-gen install uninstall run

# Default target
all: gen build test

# Build the controller binary
build:
	@echo "Building controller..."
	$(GO) build $(BUILD_FLAGS) -o $(BIN_DIR)/controller cmd/controller/main.go

# Generate client code
gen:
	@echo "Generating client code..."
	bash hack/codegen.sh
	@echo "Code generation completed!"

# Run tests
test:
	@echo "Running tests..."
	$(GO) test ./... -v

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

# Run linter
lint:
	@echo "Running linter..."
	$(GOLINT) run

# Verify generated code is up to date
verify-gen: gen
	@echo "Verifying generated code is up to date..."
	git diff --exit-code -- pkg/generated pkg/apis/*/*/zz_generated.*.go || (echo "Generated code is out of date, please run 'make gen'" && exit 1)

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	rm -rf $(BIN_DIR)
	go clean -cache

# Install CRDs into the cluster
install:
	@echo "Installing CRDs..."
	kubectl apply -f config/crds

# Uninstall CRDs from the cluster
uninstall:
	@echo "Uninstalling CRDs..."
	kubectl delete -f config/crds

# Run the controller locally
run: build
	@echo "Running controller locally..."
	$(BIN_DIR)/controller --kubeconfig=$(KUBECONFIG)

# Print help information
help:
	@echo "Available targets:"
	@echo "  all          - Generate code, build, and run tests (default)"
	@echo "  build        - Build the controller binary"
	@echo "  gen          - Generate client code using codegen.sh"
	@echo "  test         - Run tests"
	@echo "  fmt          - Format code"
	@echo "  lint         - Run linter"
	@echo "  verify-gen   - Verify generated code is up to date"
	@echo "  clean        - Clean build artifacts"
	@echo "  install      - Install CRDs into the cluster"
	@echo "  uninstall    - Uninstall CRDs from the cluster"
	@echo "  run          - Run the controller locally"
	@echo "  help         - Print this help information"
