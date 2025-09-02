#!/bin/bash
# Creates a proper macOS .app bundle for Board Game Library

APP_NAME="Board Game Library"
BUNDLE_NAME="BoardGameLibrary.app"
BUNDLE_DIR="$BUNDLE_NAME/Contents"

echo "Creating macOS app bundle..."

# Clean previous bundle
rm -rf "$BUNDLE_NAME"

# Create bundle structure
mkdir -p "$BUNDLE_DIR/MacOS"
mkdir -p "$BUNDLE_DIR/Resources"

# Build the executable
echo "Building executable..."
go build -ldflags="-s -w" -o "$BUNDLE_DIR/MacOS/BoardGameLibrary" ./cmd/server

# Make executable
chmod +x "$BUNDLE_DIR/MacOS/BoardGameLibrary"

# Copy resources
cp -r web "$BUNDLE_DIR/Resources/" 2>/dev/null || true
cp -r docs "$BUNDLE_DIR/Resources/" 2>/dev/null || true

# Create Info.plist
cat > "$BUNDLE_DIR/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>BoardGameLibrary</string>
    <key>CFBundleIdentifier</key>
    <string>com.boardgamelibrary.app</string>
    <key>CFBundleName</key>
    <string>$APP_NAME</string>
    <key>CFBundleVersion</key>
    <string>1.0.0</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0.0</string>
    <key>CFBundleInfoDictionaryVersion</key>
    <string>6.0</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleSignature</key>
    <string>????</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.12</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>LSUIElement</key>
    <true/>
</dict>
</plist>
EOF

# Create launcher script that sets up environment
cat > "$BUNDLE_DIR/MacOS/launcher.sh" << 'EOF'
#!/bin/bash
# Get the directory where the app bundle is located
BUNDLE_DIR="$(dirname "$0")/.."
RESOURCES_DIR="$BUNDLE_DIR/Resources"

# Create application directory in user's Application Support
mkdir -p "$HOME/Library/Application Support/BoardGameLibrary"

# Set environment variables
export DATABASE_PATH="$HOME/Library/Application Support/BoardGameLibrary/library.db"
export SERVER_PORT=8080
export LOG_LEVEL=info

# Change to resources directory so relative paths work
cd "$RESOURCES_DIR"

# Start the application
exec "$BUNDLE_DIR/MacOS/BoardGameLibrary"
EOF

chmod +x "$BUNDLE_DIR/MacOS/launcher.sh"

echo ""
echo "macOS app bundle created: $BUNDLE_NAME"
echo "You can now:"
echo "1. Double-click the .app to run"
echo "2. Move it to /Applications"
echo "3. Distribute as a .dmg or .zip"
echo ""
echo "Note: The app will store data in ~/Library/Application Support/BoardGameLibrary/"