package database

import (
	"os"
	"path/filepath"
	"runtime"
)

// GetDefaultDatabasePath returns the best default database path for the current OS
// without requiring admin privileges
func GetDefaultDatabasePath() string {
	switch runtime.GOOS {
	case "windows":
		return getWindowsUserPath()
	case "darwin":
		return getMacOSUserPath()
	case "linux":
		return getLinuxUserPath()
	default:
		// Fallback to current directory
		return "./data/library.db"
	}
}

// getWindowsUserPath returns a Windows-specific path in user's AppData
func getWindowsUserPath() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		// Fallback to current directory if APPDATA is not available
		return "./data/library.db"
	}
	return filepath.Join(appData, "BoardGameLibrary", "library.db")
}

// getMacOSUserPath returns a macOS-specific path in user's Application Support
func getMacOSUserPath() string {
	home := os.Getenv("HOME")
	if home == "" {
		return "./data/library.db"
	}
	return filepath.Join(home, "Library", "Application Support", "BoardGameLibrary", "library.db")
}

// getLinuxUserPath returns a Linux-specific path following XDG Base Directory Specification
func getLinuxUserPath() string {
	xdgData := os.Getenv("XDG_DATA_HOME")
	if xdgData != "" {
		return filepath.Join(xdgData, "board-game-library", "library.db")
	}
	
	home := os.Getenv("HOME")
	if home == "" {
		return "./data/library.db"
	}
	return filepath.Join(home, ".local", "share", "board-game-library", "library.db")
}

// EnsureDirectoryExists creates the directory for the database path if it doesn't exist
func EnsureDirectoryExists(dbPath string) error {
	dir := filepath.Dir(dbPath)
	return os.MkdirAll(dir, 0755)
}