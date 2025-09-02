#!/bin/bash
# Board Game Library Launcher for macOS
# This script ensures the database is created in user's Application Support directory

# Create application directory in user's Application Support (no admin rights needed)
mkdir -p "$HOME/Library/Application Support/BoardGameLibrary"

# Set environment variables
export DATABASE_PATH="$HOME/Library/Application Support/BoardGameLibrary/library.db"
export SERVER_PORT=8080
export LOG_LEVEL=info

# Display startup information
echo "Starting Board Game Library..."
echo "Database location: $DATABASE_PATH"
echo "Server will run on: http://localhost:$SERVER_PORT"
echo ""

# Start the application
./board-game-library

# Check exit status
if [ $? -ne 0 ]; then
    echo ""
    echo "Application exited with error. Press Enter to close..."
    read
fi