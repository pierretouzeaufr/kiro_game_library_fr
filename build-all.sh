#!/bin/bash
# Universal build script for Board Game Library

echo "🎮 Board Game Library - Universal Build Script"
echo "=============================================="
echo ""

# Detect current OS
OS=$(uname -s)
case "$OS" in
    Darwin)
        CURRENT_OS="macOS"
        ;;
    Linux)
        CURRENT_OS="Linux"
        ;;
    MINGW*|CYGWIN*|MSYS*)
        CURRENT_OS="Windows"
        ;;
    *)
        CURRENT_OS="Unknown"
        ;;
esac

echo "Detected OS: $CURRENT_OS"
echo ""

# Function to build for specific platform
build_platform() {
    local platform=$1
    echo "📦 Building for $platform..."
    
    case "$platform" in
        "windows")
            if command -v cmd.exe >/dev/null 2>&1; then
                cmd.exe /c "build\\windows\\build-windows-dist.bat"
            else
                echo "⚠️  Windows build tools not available on this system"
                return 1
            fi
            ;;
        "macos")
            if [ -f "build/macos/build-macos-dist.sh" ]; then
                chmod +x build/macos/build-macos-dist.sh
                ./build/macos/build-macos-dist.sh
            else
                echo "⚠️  macOS build script not found"
                return 1
            fi
            ;;
        "linux")
            make build-linux
            ;;
        *)
            echo "❌ Unknown platform: $platform"
            return 1
            ;;
    esac
}

# Parse command line arguments
if [ $# -eq 0 ]; then
    echo "Usage: $0 [windows|macos|linux|all|current]"
    echo ""
    echo "Options:"
    echo "  windows  - Build Windows distribution"
    echo "  macos    - Build macOS distribution"
    echo "  linux    - Build Linux distribution"
    echo "  all      - Build all distributions"
    echo "  current  - Build for current OS ($CURRENT_OS)"
    echo ""
    exit 1
fi

case "$1" in
    "windows")
        build_platform "windows"
        ;;
    "macos")
        build_platform "macos"
        ;;
    "linux")
        build_platform "linux"
        ;;
    "all")
        echo "🌍 Building for all platforms..."
        echo ""
        build_platform "windows"
        echo ""
        build_platform "macos"
        echo ""
        build_platform "linux"
        echo ""
        echo "✅ All builds completed!"
        ;;
    "current")
        case "$CURRENT_OS" in
            "macOS")
                build_platform "macos"
                ;;
            "Linux")
                build_platform "linux"
                ;;
            "Windows")
                build_platform "windows"
                ;;
            *)
                echo "❌ Unsupported OS: $CURRENT_OS"
                exit 1
                ;;
        esac
        ;;
    *)
        echo "❌ Unknown option: $1"
        echo "Use: $0 [windows|macos|linux|all|current]"
        exit 1
        ;;
esac