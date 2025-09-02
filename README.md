# Board Game Library Management System

[![Tests](https://github.com/YOUR_USERNAME/YOUR_REPO_NAME/workflows/Tests/badge.svg)](https://github.com/YOUR_USERNAME/YOUR_REPO_NAME/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/YOUR_USERNAME/YOUR_REPO_NAME)](https://goreportcard.com/report/github.com/YOUR_USERNAME/YOUR_REPO_NAME)

A Go-based web application for managing a board game library with user borrowing capabilities.

## Features

- User management and registration
- Game inventory management
- Borrowing and return workflow
- Overdue alerts and notifications
- Responsive web interface with HTMX
- SQLite database for local storage
- Cross-platform support (Windows, macOS, Linux)
- No administrator rights required

## Project Structure

```
board-game-library/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── models/          # Data models
│   ├── repositories/    # Data access layer
│   ├── services/        # Business logic
│   └── handlers/        # HTTP handlers
├── pkg/
│   └── database/        # Database utilities
├── web/
│   ├── templates/       # HTML templates
│   └── static/          # CSS, JS, images
└── go.mod
```

## Dependencies

- **Gin**: Web framework for HTTP routing and middleware
- **SQLite3**: Database driver for local data storage
- **HTMX**: Frontend interactivity without complex JavaScript
- **Tailwind CSS**: Utility-first CSS framework

## Quick Start (Development)

### Option 1: Native Go
1. Clone the repository
2. Install dependencies: `go mod tidy`
3. Run the application: `go run cmd/server/main.go`
4. Open http://localhost:8080 in your browser

### Option 2: Docker
```bash
# Production mode
docker-compose up -d

# Development mode (with hot reload)
docker-compose -f docker-compose.dev.yml up

# Build and run
docker-compose up --build
```

## Building for Distribution

### 🪟 Windows

#### Option 1: Simple Build
```bash
# From project root
./build/windows/build.bat
```

#### Option 2: Complete Distribution
```bash
# Creates a full Windows distribution with launchers
./build/windows/build-windows-dist.bat
```

**What you get:**
- `windows-dist/board-game-library.exe` - Main executable
- `windows-dist/run.bat` - Standard launcher (stores data in AppData)
- `windows-dist/run-portable.bat` - Portable launcher (stores data locally)
- `windows-dist/config.env` - Configuration file
- `windows-dist/web/` - Web assets
- `windows-dist/docs/` - API documentation

**How to run:**
1. Extract the `windows-dist` folder
2. Double-click `run.bat` (recommended) or `run-portable.bat`
3. Open http://localhost:8080 in your browser

### 🍎 macOS

#### Option 1: Simple Build
```bash
# From project root
./build/macos/build.sh
```

#### Option 2: Complete Distribution
```bash
# Creates a full macOS distribution with launchers
./build/macos/build-macos-dist.sh
```

#### Option 3: Native App Bundle
```bash
# Creates a native .app bundle
./build/macos/create-app-bundle.sh
```

**What you get:**
- `macos-dist/board-game-library` - Main executable
- `macos-dist/run.sh` - Standard launcher (stores data in Application Support)
- `macos-dist/run-portable.sh` - Portable launcher (stores data locally)
- `macos-dist/config.env` - Configuration file
- `macos-dist/web/` - Web assets
- `macos-dist/docs/` - API documentation

**How to run:**
1. Extract the `macos-dist` folder
2. Run `./run.sh` (recommended) or `./run-portable.sh`
3. Open http://localhost:8080 in your browser

### 🐧 Linux

```bash
# Build Linux distribution
make build-linux
```

### 🚀 Universal Build Script

```bash
# Build for current OS
./build-all.sh current

# Build for specific platform
./build-all.sh windows
./build-all.sh macos
./build-all.sh linux

# Build for all platforms
./build-all.sh all
```

### 🛠️ Using Make (Alternative)

```bash
# Build distributions for all platforms
make build-all

# Build specific platform
make build-windows
make build-macos
make build-linux
```

## Data Storage Locations

### Standard Mode (Recommended)
- **Windows**: `%APPDATA%\BoardGameLibrary\library.db`
- **macOS**: `~/Library/Application Support/BoardGameLibrary/library.db`
- **Linux**: `~/.local/share/board-game-library/library.db`

### Portable Mode
- **All platforms**: `./data/library.db` (next to executable)

## Configuration

Edit the `config.env` file to customize:
- Server port (default: 8080)
- Database path
- Logging level
- Alert settings

Example:
```env
SERVER_PORT=8080
DATABASE_PATH=./data/library.db
LOG_LEVEL=info
```

## Development

The application follows a clean architecture pattern with clear separation of concerns:

- **Models**: Define data structures and validation
- **Repositories**: Handle data persistence and retrieval
- **Services**: Implement business logic and workflows
- **Handlers**: Process HTTP requests and responses

## Requirements

### For Development
- Go 1.21 or higher
- SQLite3 (included with go-sqlite3 driver)

### For End Users
- **Windows**: Windows 7 or higher (no additional requirements)
- **macOS**: macOS 10.12 or higher (no additional requirements)
- **Linux**: Any modern Linux distribution (no additional requirements)

## Available Make Targets

```bash
make help              # Show all available targets
make build-windows     # Build Windows distribution
make build-macos       # Build macOS distribution  
make build-linux       # Build Linux distribution
make build-all         # Build all distributions
make app-bundle        # Create macOS .app bundle
make clean             # Clean build artifacts
make test              # Run tests
make dev-run           # Start development server
make docker-build      # Build Docker image
make docker-up         # Start with Docker Compose
make docker-dev        # Start development with Docker
```

## Continuous Integration

The project includes GitHub Actions workflows for automated testing:

### 🔄 Automated Tests (`tests.yml`)
- Runs on every push to `main` or `develop` branches
- Runs on pull requests
- Includes: tests, linting, security scanning
- Generates coverage reports

### 🧪 Simple Tests (`simple-tests.yml`)  
- Manual trigger only (workflow_dispatch)
- Basic compilation and test verification
- Lightweight alternative for quick checks

### Running Tests Locally
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test -v ./pkg/database/...
go test -v ./internal/services/...
```

## Docker Usage

### Production Deployment
```bash
# Start the application
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the application
docker-compose down

# Update and restart
docker-compose pull && docker-compose up -d
```

### Development with Docker
```bash
# Start development environment
docker-compose -f docker-compose.dev.yml up

# Rebuild after changes
docker-compose -f docker-compose.dev.yml up --build
```

### Docker Data Persistence
- Database is stored in Docker volume `board-game-data`
- To backup: `docker run --rm -v board-game-data:/data -v $(pwd):/backup alpine tar czf /backup/backup.tar.gz -C /data .`
- To restore: `docker run --rm -v board-game-data:/data -v $(pwd):/backup alpine tar xzf /backup/backup.tar.gz -C /data`

## Troubleshooting

### Windows
- If Windows Defender blocks the app, click "More info" → "Run anyway"
- No administrator rights required

### macOS  
- If macOS blocks the app, go to System Preferences → Security & Privacy → Click "Allow Anyway"
- For the .app bundle, you may need to right-click → Open the first time

### Docker
- Make sure Docker is running
- Check port 8080 is not in use: `docker ps` or `lsof -i :8080`
- View container logs: `docker-compose logs board-game-library`

### All Platforms
- Make sure port 8080 is not in use by another application
- Check the console output for any error messages
- Database is created automatically on first run