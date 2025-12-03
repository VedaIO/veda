# Makefile for procguard

# Use git describe to get a version string.
# Example: v1.0.0-3-g1234567
# Fallback to 'dev' if not in a git repository.
VERSION ?= $(shell git describe --tags --always --dirty --first-parent 2>/dev/null || echo "dev")

.PHONY: all build dev fmt clean

all: build

build:
	@echo "Building ProcGuard for windows..."
	cd wails-app/frontend && npm run build
	cd wails-app && CGO_ENABLED=1 CC="zig cc -target x86_64-windows-gnu" GOOS=windows GOARCH=amd64 go build -tags desktop,production -ldflags="-w -s -X main.version=$(VERSION) -H windowsgui" -o build/bin/ProcGuard.exe

build-debug:
	@echo "Building ProcGuard for windows (debug)..."
	cd wails-app/frontend && npm run build
	cd wails-app && CGO_ENABLED=1 CC="zig cc -target x86_64-windows-gnu" GOOS=windows GOARCH=amd64 go build -tags desktop,production -ldflags="-X main.version=$(VERSION)" -o build/bin/ProcGuard.exe

fmt:
	@echo "Formatting code..."
	cd wails-app && go fmt ./...
	cd wails-app/frontend && npm run format

lint:
	cd wails-app && GOOS=windows golangci-lint run
	cd wails-app/frontend && npm run lint

clean:
	@echo "Cleaning..."
	rm -rf wails-app/build/bin
	rm -rf wails-app/frontend/dist
	rm -rf wails-app/frontend/wailsjs
