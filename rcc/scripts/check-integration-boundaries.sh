#!/bin/bash
# CI guard script to ensure integration tests maintain proper boundaries
# This script fails if integration tests violate architectural boundaries

set -e

echo "🔍 Checking integration test boundaries..."

# Check for HTTP/SSE usage in integration tests
echo "  Checking for HTTP/SSE usage..."
HTTP_VIOLATIONS=$(find test/integration -name "*.go" -exec grep -l "net/http\|httptest\|ResponseRecorder\|http\.Client" {} \; 2>/dev/null || true)
if [ -n "$HTTP_VIOLATIONS" ]; then
    echo "❌ HTTP/SSE violations found in integration tests:"
    echo "$HTTP_VIOLATIONS"
    echo "Integration tests should not use HTTP - move to test/e2e/"
    exit 1
fi

# Check for filesystem access in integration tests
echo "  Checking for filesystem access..."
FS_VIOLATIONS=$(find test/integration -name "*.go" -exec grep -l "os\.Read\|os\.Open\|ioutil\.Read\|os\.Write\|os\.Create" {} \; 2>/dev/null || true)
if [ -n "$FS_VIOLATIONS" ]; then
    echo "❌ Filesystem access violations found in integration tests:"
    echo "$FS_VIOLATIONS"
    echo "Integration tests should use mocks instead of filesystem access"
    exit 1
fi

# Check for long sleeps in integration tests
echo "  Checking for long sleeps..."
SLEEP_VIOLATIONS=$(find test/integration -name "*.go" -exec grep -l "time\.Sleep.*[6-9][0-9][0-9]\|time\.Sleep.*[0-9][0-9][0-9][0-9]" {} \; 2>/dev/null || true)
if [ -n "$SLEEP_VIOLATIONS" ]; then
    echo "❌ Long sleep violations found in integration tests:"
    echo "$SLEEP_VIOLATIONS"
    echo "Integration tests should use fast timeouts (< 100ms)"
    exit 1
fi

# Check for "DRIFT" logs that should be test failures
echo "  Checking for DRIFT logs..."
DRIFT_VIOLATIONS=$(find test/integration -name "*.go" -exec grep -l "DRIFT\|drift" {} \; 2>/dev/null || true)
if [ -n "$DRIFT_VIOLATIONS" ]; then
    echo "❌ DRIFT log violations found in integration tests:"
    echo "$DRIFT_VIOLATIONS"
    echo "Replace DRIFT logs with proper test failures for real bugs"
    exit 1
fi

echo "✅ All integration test boundaries respected!"
echo "   - No HTTP/SSE usage"
echo "   - No filesystem access"
echo "   - No long sleeps"
echo "   - No DRIFT logs"


