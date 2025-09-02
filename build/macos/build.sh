#!/bin/bash
echo "Building board-game-library for macOS..."

# Create distribution directory
mkdir -p dist/data
mkdir -p dist/web
mkdir -p dist/docs

# Build the executable
go build -o dist/board-game-library ./cmd/server

# Make executable
chmod +x dist/board-game-library

# Copy web assets and docs
cp -r web/* dist/web/ 2>/dev/null || true
cp -r docs/* dist/docs/ 2>/dev/null || true

# Create example config file
cp config.example.env dist/config.env

echo "Build complete. Distribution files are in the 'dist' directory."
echo ""
echo "To run:"
echo "1. Navigate to the dist directory"
echo "2. Edit config.env if needed (database will be created in data/ folder)"
echo "3. Run: ./board-game-library"