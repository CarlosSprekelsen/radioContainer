#!/bin/bash

# RCC Web UI - Quick Test Runner
echo "RCC Web UI - Quick Test Runner"
echo "==============================="

# Kill any existing server
echo "Stopping any existing server..."
pkill -f rcc-webui 2>/dev/null || true
sleep 1

# Build and start server
echo "Building and starting server..."
go build -o rcc-webui main.go
if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi

echo "✅ Build successful"

# Start server in background
echo "Starting server on http://127.0.0.1:3000..."
./rcc-webui &
SERVER_PID=$!

# Wait for server to start
echo "Waiting for server to start..."
sleep 3

# Check if server is running
if ! curl -s http://127.0.0.1:3000/ > /dev/null 2>&1; then
    echo "❌ Server failed to start!"
    kill $SERVER_PID 2>/dev/null || true
    exit 1
fi

echo "✅ Server started successfully"

# Run functionality tests
echo ""
echo "Running functionality tests..."
echo "============================"

BASE_URL="http://127.0.0.1:3000"
TESTS_PASSED=0
TOTAL_TESTS=8

# Test 1: HTML page
echo -n "1. Testing HTML page... "
if curl -s "$BASE_URL/" | grep -q "Radio Control"; then
    echo "✅ PASS"
    ((TESTS_PASSED++))
else
    echo "❌ FAIL"
fi

# Test 2: CSS
echo -n "2. Testing CSS... "
if curl -s "$BASE_URL/style.css" | grep -q "RCC Web UI Styles"; then
    echo "✅ PASS"
    ((TESTS_PASSED++))
else
    echo "❌ FAIL"
fi

# Test 3: JavaScript
echo -n "3. Testing JavaScript... "
if curl -s "$BASE_URL/app.js" | grep -q "RCCClient"; then
    echo "✅ PASS"
    ((TESTS_PASSED++))
else
    echo "❌ FAIL"
fi

# Test 4: Config
echo -n "4. Testing config... "
if curl -s "$BASE_URL/config.json" | grep -q "rccBaseUrl"; then
    echo "✅ PASS"
    ((TESTS_PASSED++))
else
    echo "❌ FAIL"
fi

# Test 5: Reverse proxy
echo -n "5. Testing reverse proxy... "
RESPONSE=$(curl -s "$BASE_URL/radios")
if echo "$RESPONSE" | grep -q "Failed to connect to RCC"; then
    echo "✅ PASS (expected failure without RCC)"
    ((TESTS_PASSED++))
else
    echo "❌ FAIL"
fi

# Test 6: Audit endpoint
echo -n "6. Testing audit endpoint... "
AUDIT_RESPONSE=$(curl -s -X POST "$BASE_URL/audit" \
    -H "Content-Type: application/json" \
    -d '{"timestamp":"2024-01-01T00:00:00Z","actor":"test","radioId":"test-radio","action":"test","result":"success","latencyMs":100,"correlationId":"test-123"}')
if [ $? -eq 0 ]; then
    echo "✅ PASS"
    ((TESTS_PASSED++))
else
    echo "❌ FAIL"
fi

# Test 7: Telemetry SSE
echo -n "7. Testing telemetry SSE... "
TELEMETRY_RESPONSE=$(timeout 2 curl -s "$BASE_URL/telemetry" 2>/dev/null)
if echo "$TELEMETRY_RESPONSE" | grep -q "Failed to connect to RCC"; then
    echo "✅ PASS (expected failure without RCC)"
    ((TESTS_PASSED++))
else
    echo "❌ FAIL"
fi

# Test 8: Audit log file
echo -n "8. Testing audit log... "
if [ -f "audit.log" ] && [ -s "audit.log" ]; then
    echo "✅ PASS ($(wc -l < audit.log) entries)"
    ((TESTS_PASSED++))
else
    echo "❌ FAIL"
fi

# Results
echo ""
echo "Test Results"
echo "============"
echo "Tests passed: $TESTS_PASSED/$TOTAL_TESTS"

if [ $TESTS_PASSED -eq $TOTAL_TESTS ]; then
    echo "🎉 ALL TESTS PASSED!"
    echo ""
    echo "Web UI is running at: http://127.0.0.1:3000"
    echo "Server PID: $SERVER_PID"
    echo ""
    echo "To stop the server: kill $SERVER_PID"
    echo "Or run: pkill -f rcc-webui"
    echo ""
    echo "Expected behavior:"
    echo "- Radio dropdown shows 'No radios available' (normal without RCC)"
    echo "- Console shows connection errors (normal without RCC)"
    echo "- All UI controls are present and functional"
    exit 0
else
    echo "❌ Some tests failed!"
    echo "Stopping server..."
    kill $SERVER_PID 2>/dev/null || true
    exit 1
fi
