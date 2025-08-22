@echo off
echo Building board-game-library for Windows...
go build -o board-game-library.exe ./cmd/server
echo Build complete. You can find the executable in the current directory.
