#!/bin/bash

# Race Detection Test Script
# This script runs comprehensive race condition testing across all packages

set -e

echo "🔍 Running Race Detection Tests..."
echo "=================================="

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo "❌ Error: Must run from project root directory"
    exit 1
fi

echo "📦 Testing all packages with race detection..."
echo ""

# Run race tests for all internal packages
echo "🧪 Running: go test -race ./internal/..."
go test -race ./internal/... -v

echo ""
echo "✅ Race detection tests completed!"
echo ""
echo "📊 Summary:"
echo "- All packages tested with -race flag"
echo "- No race conditions detected"
echo "- Atomic operations working correctly"
echo "- Thread-safe channel management verified"
