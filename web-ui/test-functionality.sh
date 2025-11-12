#!/bin/bash

# RCC Web UI Functionality Test Script
echo "RCC Web UI Functionality Test"
echo "============================="

BASE_URL="http://127.0.0.1:3000"

echo "1. Testing static file serving..."
if curl -s "$BASE_URL/" | grep -q "Radio Control"; then
    echo "✓ HTML page loads correctly"
else
    echo "✗ HTML page failed to load"
    exit 1
fi

echo "2. Testing CSS serving..."
if curl -s "$BASE_URL/style.css" | grep -q "RCC Web UI Styles"; then
    echo "✓ CSS loads correctly"
else
    echo "✗ CSS failed to load"
    exit 1
fi

echo "3. Testing JavaScript serving..."
if curl -s "$BASE_URL/app.js" | grep -q "RCCClient"; then
    echo "✓ JavaScript loads correctly"
else
    echo "✗ JavaScript failed to load"
    exit 1
fi

echo "4. Testing config.json serving..."
if curl -s "$BASE_URL/config.json" | grep -q "rccBaseUrl"; then
    echo "✓ Config loads correctly"
else
    echo "✗ Config failed to load"
    exit 1
fi

echo "5. Testing reverse proxy (expected to fail without RCC)..."
RESPONSE=$(curl -s "$BASE_URL/radios")
if echo "$RESPONSE" | grep -q "Failed to connect to RCC"; then
    echo "✓ Reverse proxy working (expected failure without RCC container)"
else
    echo "✗ Reverse proxy not working correctly"
    exit 1
fi

echo "6. Testing audit endpoint..."
AUDIT_RESPONSE=$(curl -s -X POST "$BASE_URL/audit" \
    -H "Content-Type: application/json" \
    -d '{"timestamp":"2024-01-01T00:00:00Z","actor":"test","radioId":"test-radio","action":"test","result":"success","latencyMs":100,"correlationId":"test-123"}')
if [ $? -eq 0 ]; then
    echo "✓ Audit endpoint working"
else
    echo "✗ Audit endpoint failed"
    exit 1
fi

echo "7. Testing telemetry SSE endpoint (expected to fail without RCC)..."
TELEMETRY_RESPONSE=$(timeout 2 curl -s "$BASE_URL/telemetry" 2>/dev/null)
if echo "$TELEMETRY_RESPONSE" | grep -q "Failed to connect to RCC"; then
    echo "✓ Telemetry endpoint working (expected failure without RCC container)"
else
    echo "✗ Telemetry endpoint not working correctly"
    exit 1
fi

echo "8. Verifying audit log creation..."
if [ -f "audit.log" ] && [ -s "audit.log" ]; then
    echo "✓ Audit log file created and contains data"
    echo "   Audit log entries: $(wc -l < audit.log)"
else
    echo "✗ Audit log not created or empty"
    exit 1
fi

echo ""
echo "🎉 All tests passed!"
echo ""
echo "Web UI is running successfully at: $BASE_URL"
echo "Expected behavior without RCC container:"
echo "  - Radio dropdown shows 'No radios available'"
echo "  - Console shows connection errors (normal)"
echo "  - UI remains functional for testing"
echo ""
echo "To test with RCC container:"
echo "  1. Start RCC container on http://localhost:8080"
echo "  2. Refresh browser at $BASE_URL"
echo "  3. UI should connect and show available radios"
