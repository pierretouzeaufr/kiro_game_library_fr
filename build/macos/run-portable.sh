#!/bin/bash
# Board Game Library - Portable Version for macOS
# Runs entirely from the current directory (no admin rights needed)

# Create data directory in current folder
mkdir -p data

# Set environment variables for portable mode
export DATABASE_PATH="./data/library.db"
export SERVER_PORT=8080
export LOG_LEVEL=info

# Display startup information
echo "Starting Board Game Library (Portable Mode)..."
echo "Database location: $(pwd)/data/library.db"
echo "Server will run on: http://localhost:$SERVER_PORT"
echo ""
echo "Note: All data is stored in the 'data' folder next to this executable."
echo "You can move this entire folder anywhere without losing your data."
echo ""

# Start the application
./board-game-library

# Check exit status
if [ $? -ne 0 ]; then
    echo ""
    echo "Application exited with error. Press Enter to close..."
    read
fi