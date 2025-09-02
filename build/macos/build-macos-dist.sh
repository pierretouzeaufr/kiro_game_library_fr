#!/bin/bash
echo "Creating macOS distribution for Board Game Library..."

# Clean previous build
rm -rf macos-dist

# Create distribution structure
mkdir -p macos-dist/data
mkdir -p macos-dist/web
mkdir -p macos-dist/docs

# Build the executable with macOS optimizations
echo "Building executable..."
export CGO_ENABLED=1
export GOOS=darwin
export GOARCH=amd64
go build -ldflags="-s -w" -o macos-dist/board-game-library ./cmd/server

# Make the executable... executable
chmod +x macos-dist/board-game-library

# Copy required assets
echo "Copying assets..."
cp -r web/* macos-dist/web/ 2>/dev/null || true
cp -r docs/* macos-dist/docs/ 2>/dev/null || true

# Copy configuration files and launchers
cp build/macos/config.macos.env macos-dist/config.env
cp build/macos/run.sh macos-dist/
cp build/macos/run-portable.sh macos-dist/

# Make scripts executable
chmod +x macos-dist/run.sh
chmod +x macos-dist/run-portable.sh

# Create README for macOS users
cat > macos-dist/README.txt << 'EOF'
Board Game Library - macOS Distribution
======================================

Quick Start:
1. Choose your preferred mode:
   - ./run.sh: Stores data in ~/Library/Application Support (recommended)
   - ./run-portable.sh: Stores data in local folder (portable)
2. Open your browser to http://localhost:8080
3. No administrator rights required!

Configuration:
- Edit config.env to change settings
- Standard mode database: ~/Library/Application Support/BoardGameLibrary/library.db
- Portable mode database: data/library.db
- Web assets: web/ folder
- API documentation: docs/ folder

Manual Start:
If you prefer to run manually, use: ./board-game-library
Make sure to set DATABASE_PATH environment variable appropriately

Troubleshooting:
- If you get "permission denied", run: chmod +x board-game-library
- If macOS blocks the app, go to System Preferences > Security & Privacy
  and click "Allow Anyway" next to the blocked app notification

Support:
Check the documentation in the docs folder for API details.
EOF

echo ""
echo "macOS distribution created successfully in 'macos-dist' folder!"
echo ""
echo "Contents:"
echo "- board-game-library      (main executable)"
echo "- run.sh                  (standard launcher - uses Application Support)"
echo "- run-portable.sh         (portable launcher - uses local folder)"
echo "- config.env              (configuration file)"
echo "- data/                   (database directory for portable mode)"
echo "- web/                    (web assets)"
echo "- docs/                   (API documentation)"
echo "- README.txt              (instructions)"
echo ""
echo "To distribute: tar -czf board-game-library-macos.tar.gz macos-dist/"