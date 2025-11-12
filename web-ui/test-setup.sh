#!/bin/bash

# RCC Web UI Test Setup Script
# This script helps test the Web UI implementation

echo "RCC Web UI Test Setup"
echo "====================="

# Check if Go is available
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed or not in PATH"
    exit 1
fi

echo "✓ Go is available"

# Check if config.json exists
if [ ! -f "config.json" ]; then
    echo "Error: config.json not found"
    exit 1
fi

echo "✓ config.json found"

# Check if static files exist
if [ ! -f "static/index.html" ] || [ ! -f "static/style.css" ] || [ ! -f "static/app.js" ]; then
    echo "Error: Static files not found"
    exit 1
fi

echo "✓ Static files found"

# Build the application
echo "Building application..."
if go build -o rcc-webui main.go; then
    echo "✓ Application built successfully"
else
    echo "Error: Failed to build application"
    exit 1
fi

echo ""
echo "Setup complete! To run the Web UI:"
echo "1. Ensure RCC container is running on http://localhost:8080"
echo "2. Run: ./rcc-webui"
echo "3. Open browser to: http://127.0.0.1:3000"
echo ""
echo "To test without RCC container:"
echo "1. Run: ./rcc-webui"
echo "2. Check browser console for connection errors (expected)"
echo "3. UI will show 'No radios available' (expected behavior)"
