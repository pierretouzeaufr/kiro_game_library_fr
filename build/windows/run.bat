@echo off
REM Board Game Library Launcher for Windows
REM This script ensures the database is created in a data directory

REM Create application directory in user's AppData (no admin rights needed)
if not exist "%APPDATA%\BoardGameLibrary" mkdir "%APPDATA%\BoardGameLibrary"

REM Set environment variables
set DATABASE_PATH=%APPDATA%\BoardGameLibrary\library.db
set SERVER_PORT=8080
set LOG_LEVEL=info

REM Display startup information
echo Starting Board Game Library...
echo Database will be stored in: %DATABASE_PATH%
echo Server will run on port: %SERVER_PORT%
echo.

REM Start the application
board-game-library.exe

REM Keep window open if there's an error
if errorlevel 1 (
    echo.
    echo Application exited with error. Press any key to close...
    pause >nul
)