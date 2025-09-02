@echo off
REM Board Game Library - Portable Version
REM Runs entirely from the current directory (no admin rights needed)

REM Create data directory in current folder
if not exist "data" mkdir data

REM Set environment variables for portable mode
set DATABASE_PATH=data\library.db
set SERVER_PORT=8080
set LOG_LEVEL=info

REM Display startup information
echo Starting Board Game Library (Portable Mode)...
echo Database location: %CD%\data\library.db
echo Server will run on: http://localhost:%SERVER_PORT%
echo.
echo Note: All data is stored in the 'data' folder next to this executable.
echo You can move this entire folder anywhere without losing your data.
echo.

REM Start the application
board-game-library.exe

REM Keep window open if there's an error
if errorlevel 1 (
    echo.
    echo Application exited with error. Press any key to close...
    pause >nul
)