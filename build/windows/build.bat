@echo off
echo Building board-game-library for Windows...

REM Create distribution directory
if not exist "dist" mkdir dist
if not exist "dist\data" mkdir dist\data
if not exist "dist\web" mkdir dist\web
if not exist "dist\docs" mkdir dist\docs

REM Build the executable
go build -o dist\board-game-library.exe ./cmd/server

REM Copy web assets and docs
xcopy /E /I web dist\web
xcopy /E /I docs dist\docs

REM Create example config file
copy config.example.env dist\config.env

echo Build complete. Distribution files are in the 'dist' directory.
echo.
echo To run:
echo 1. Navigate to the dist directory
echo 2. Edit config.env if needed (database will be created in data\ folder)
echo 3. Run: board-game-library.exe
