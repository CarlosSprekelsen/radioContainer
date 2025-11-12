#!/bin/bash

# Performance testing scenarios using Vegeta
# Install: go install github.com/tsenart/vegeta@latest

set -e

echo "🚀 Starting Radio Control Container Performance Tests"
echo "=================================================="

# Check if vegeta is installed
if ! command -v vegeta &> /dev/null; then
    echo "❌ Vegeta not found. Install with: go install github.com/tsenart/vegeta@latest"
    exit 1
fi

# Start the server in background (you'll need to implement this)
echo "📡 Starting test server..."
# For now, we'll assume the server is running on localhost:8080
# In a real scenario, you'd start the server here

SERVER_URL="http://localhost:8080"
echo "🔗 Server URL: $SERVER_URL"

# Wait for server to be ready
echo "⏳ Waiting for server to be ready..."
sleep 2

# Test 1: List radios endpoint
echo ""
echo "📊 Test 1: List Radios (100 req/s for 30s)"
echo "----------------------------------------"
echo "GET $SERVER_URL/api/v1/radios" | vegeta attack -duration=30s -rate=100 | vegeta report

# Test 2: Set power endpoint
echo ""
echo "📊 Test 2: Set Power (50 req/s for 30s)"
echo "------------------------------------"
echo '{"method":"POST","url":"'$SERVER_URL'/api/v1/radios/silvus-001/power","body":"{\"powerDbm\":20}","header":{"Content-Type":["application/json"]}}' | vegeta attack -duration=30s -rate=50 | vegeta report

# Test 3: Set channel endpoint
echo ""
echo "📊 Test 3: Set Channel (25 req/s for 30s)"
echo "--------------------------------------"
echo '{"method":"POST","url":"'$SERVER_URL'/api/v1/radios/silvus-001/channel","body":"{\"channelIndex\":6}","header":{"Content-Type":["application/json"]}}' | vegeta attack -duration=30s -rate=25 | vegeta report

# Test 4: Mixed workload
echo ""
echo "📊 Test 4: Mixed Workload (100 req/s for 60s)"
echo "--------------------------------------------"
cat << EOF | vegeta attack -duration=60s -rate=100 | vegeta report
GET $SERVER_URL/api/v1/radios
POST $SERVER_URL/api/v1/radios/silvus-001/power
{"method":"POST","url":"$SERVER_URL/api/v1/radios/silvus-001/power","body":"{\"powerDbm\":15}","header":{"Content-Type":["application/json"]}}
POST $SERVER_URL/api/v1/radios/silvus-001/channel
{"method":"POST","url":"$SERVER_URL/api/v1/radios/silvus-001/channel","body":"{\"channelIndex\":1}","header":{"Content-Type":["application/json"]}}
EOF

# Test 5: High load test
echo ""
echo "📊 Test 5: High Load (200 req/s for 30s)"
echo "---------------------------------------"
echo "GET $SERVER_URL/api/v1/radios" | vegeta attack -duration=30s -rate=200 | vegeta report

echo ""
echo "✅ Performance testing completed!"
echo "=================================================="
echo "📈 Check the results above for:"
echo "   - Latency percentiles (p50, p95, p99)"
echo "   - Error rates"
echo "   - Throughput (req/s)"
echo "   - Memory usage patterns"