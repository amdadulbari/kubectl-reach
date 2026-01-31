# Makefile for kubectl-reach (Krew plugin).
# Follows structure from https://github.com/replicatedhq/krew-plugin-template

export GO111MODULE=on

BINARY_NAME := kubectl-reach
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS     := -ldflags "-s -w -X github.com/kubectl-reach/kubectl-reach/pkg/version.Version=$(VERSION)"

.PHONY: build bin fmt vet test lint clean dist verify ci
.PHONY: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64

# Default: build for current platform into bin/
build: fmt vet
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/plugin

# Alias for build (template convention)
bin: build

fmt:
	go fmt ./pkg/... ./cmd/...

vet:
	go vet ./pkg/... ./cmd/...

test:
	go test ./pkg/... ./cmd/... -v -count=1 -coverprofile=cover.out -covermode=atomic

# Lint with golangci-lint (install: https://golangci-lint.run/usage/install/)
lint:
	golangci-lint run ./...

# CI: fmt, vet, test, lint (run before push)
ci: fmt vet test lint

clean:
	rm -rf bin/ dist/

# Cross-build for all Krew platforms
build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/plugin
build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64 ./cmd/plugin
build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/plugin
build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/plugin
build-windows-amd64:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/plugin

# Create tarballs for Krew (binary name must be kubectl-reach in archive)
dist: build-all
	@mkdir -p dist staging
	cp bin/$(BINARY_NAME)-linux-amd64 staging/$(BINARY_NAME) && tar czf dist/$(BINARY_NAME)_$(VERSION)_linux_amd64.tar.gz -C staging $(BINARY_NAME) && rm staging/$(BINARY_NAME)
	cp bin/$(BINARY_NAME)-linux-arm64 staging/$(BINARY_NAME) && tar czf dist/$(BINARY_NAME)_$(VERSION)_linux_arm64.tar.gz -C staging $(BINARY_NAME) && rm staging/$(BINARY_NAME)
	cp bin/$(BINARY_NAME)-darwin-amd64 staging/$(BINARY_NAME) && tar czf dist/$(BINARY_NAME)_$(VERSION)_darwin_amd64.tar.gz -C staging $(BINARY_NAME) && rm staging/$(BINARY_NAME)
	cp bin/$(BINARY_NAME)-darwin-arm64 staging/$(BINARY_NAME) && tar czf dist/$(BINARY_NAME)_$(VERSION)_darwin_arm64.tar.gz -C staging $(BINARY_NAME) && rm staging/$(BINARY_NAME)
	cp bin/$(BINARY_NAME)-windows-amd64.exe staging/$(BINARY_NAME).exe && tar czf dist/$(BINARY_NAME)_$(VERSION)_windows_amd64.tar.gz -C staging $(BINARY_NAME).exe && rm staging/$(BINARY_NAME).exe
	@rmdir staging 2>/dev/null || true

# Verify (tidy + vet + test)
verify: fmt
	go mod tidy
	go vet ./pkg/... ./cmd/...
	go test ./pkg/... ./cmd/... -count=1
