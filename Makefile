# Board Game Library - Build System

.PHONY: help build-windows build-macos build-linux clean test test-all test-coverage docker-build docker-up docker-dev docker-down docker-logs

help:
	@echo "Board Game Library Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  build-windows    Build Windows distribution"
	@echo "  build-macos      Build macOS distribution"
	@echo "  build-linux      Build Linux distribution"
	@echo "  build-all        Build all distributions"
	@echo "  app-bundle       Create macOS .app bundle"
	@echo "  clean           Clean all build artifacts"
	@echo "  test            Run working tests"
	@echo "  test-all        Run all tests"
	@echo "  test-coverage   Run tests with coverage report"
	@echo "  docker-build    Build Docker image"
	@echo "  docker-up       Start with Docker Compose"
	@echo "  docker-dev      Start development with Docker"
	@echo "  docker-down     Stop Docker containers"
	@echo "  docker-logs     Show Docker logs"
	@echo ""

build-windows:
	@echo "Building Windows distribution..."
	@./build/windows/build-windows-dist.bat

build-macos:
	@echo "Building macOS distribution..."
	@./build/macos/build-macos-dist.sh

build-linux:
	@echo "Building Linux distribution..."
	@mkdir -p linux-dist/data linux-dist/web linux-dist/docs
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o linux-dist/board-game-library ./cmd/server
	@cp -r web/* linux-dist/web/ 2>/dev/null || true
	@cp -r docs/* linux-dist/docs/ 2>/dev/null || true
	@cp config.example.env linux-dist/config.env
	@chmod +x linux-dist/board-game-library
	@echo "Linux distribution created in 'linux-dist' folder"

build-all: build-windows build-macos build-linux

app-bundle:
	@echo "Creating macOS app bundle..."
	@./build/macos/create-app-bundle.sh

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf dist windows-dist macos-dist linux-dist *.app
	@echo "Clean complete"

test:
	@echo "Running tests..."
	@go test -v ./pkg/database/... ./internal/services/... ./internal/logging/... ./internal/config/...

test-all:
	@echo "Running all tests (including potentially failing ones)..."
	@go test ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./pkg/database/... ./internal/services/... ./internal/logging/... ./internal/config/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Development targets
dev-run:
	@echo "Starting development server..."
	@go run ./cmd/server

dev-build:
	@echo "Building for current platform..."
	@go build -o board-game-library ./cmd/server

# Docker targets
docker-build:
	@echo "Building Docker image..."
	@docker-compose build

docker-up:
	@echo "Starting with Docker Compose..."
	@docker-compose up -d

docker-dev:
	@echo "Starting development environment with Docker..."
	@docker-compose -f docker-compose.dev.yml up

docker-down:
	@echo "Stopping Docker containers..."
	@docker-compose down

docker-logs:
	@echo "Showing Docker logs..."
	@docker-compose logs -f