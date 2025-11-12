#!/bin/bash

# Anti-peek enforcement script for E2E tests
# Ensures E2E tests are truly black-box and don't access internal packages
# Source: Architecture §8.3a - Event-first + duty-cycled probes

set -e

: "${ANTI_PEEK_SCOPE:=e2e}"

echo "🔍 Running anti-peek enforcement checks (scope: $ANTI_PEEK_SCOPE)..."

# Check for internal package imports (excluding manifest files)
echo "Checking for internal package imports..."
INTERNAL_IMPORTS=$(grep -r "github.com/.*/internal/" test/e2e/ --include="*.go" || true)
if [ -n "$INTERNAL_IMPORTS" ]; then
    echo "❌ FAIL: Found internal package imports in E2E tests:"
    echo "$INTERNAL_IMPORTS"
    exit 1
fi

# Check for non-HTTP server symbols
echo "Checking for non-HTTP server symbols..."
SERVER_SYMBOLS=$(grep -r "\.Mux()\|\.Handler()\|server\.\*" test/e2e/ || true)
if [ -n "$SERVER_SYMBOLS" ]; then
    echo "❌ FAIL: Found non-HTTP server symbols in E2E tests:"
    echo "$SERVER_SYMBOLS"
    exit 1
fi

# Check for direct access to internal components
echo "Checking for direct access to internal components..."
INTERNAL_COMPONENTS=$(grep -r "radio\.Manager\|telemetry\.Hub\|command\.Orchestrator\|auth\.Service\|config\.Store" test/e2e/ || true)
if [ -n "$INTERNAL_COMPONENTS" ]; then
    echo "❌ FAIL: Found direct access to internal components in E2E tests:"
    echo "$INTERNAL_COMPONENTS"
    exit 1
fi

# Check for concrete adapter types
echo "Checking for concrete adapter types..."
ADAPTER_TYPES=$(grep -r "silvusmock\.\|adapter\." test/e2e/ || true)
if [ -n "$ADAPTER_TYPES" ]; then
    echo "❌ FAIL: Found concrete adapter types in E2E tests:"
    echo "$ADAPTER_TYPES"
    exit 1
fi

# Check for audit log access via harness (no audit in HTTP contract)
echo "Checking for audit log access..."
AUDIT_ACCESS=$(grep -rnE "server\\.GetAuditLogs\\(" test/e2e/ --include="*.go" || true)
if [ -n "$AUDIT_ACCESS" ]; then
    echo "❌ FAIL: Found direct audit log access in E2E tests:"
    echo "$AUDIT_ACCESS"
    echo "Audit logs are server-side only per Architecture §8.6"
    exit 1
fi

# Check for config access (should use environment variables or HTTP)
echo "Checking for config access..."
CONFIG_ACCESS=$(grep -r "config\." test/e2e/ || true)
if [ -n "$CONFIG_ACCESS" ]; then
    echo "❌ FAIL: Found direct config access in E2E tests:"
    echo "$CONFIG_ACCESS"
    echo "Use environment variables or HTTP endpoints for config"
    exit 1
fi

# Check for allowed imports only
echo "Checking for allowed imports only..."
ALLOWED_IMPORTS="net/http|net/http/httptest|encoding/json|testing|time|context|strings|os|path/filepath|github.com/radio-control/rcc/test/harness|github.com/radio-control/rcc/test/fixtures"
FORBIDDEN_IMPORTS=$(grep -r "github.com" test/e2e/ | grep -v -E "($ALLOWED_IMPORTS)" || true)
if [ -n "$FORBIDDEN_IMPORTS" ]; then
    echo "❌ FAIL: Found forbidden imports in E2E tests:"
    echo "$FORBIDDEN_IMPORTS"
    exit 1
fi

# Check for timing literals (should use config or fixtures)
echo "Checking for timing literals..."
TIMING_LITERALS=$(grep -r "time\.Sleep\|time\.Duration.*[0-9]" test/e2e/ | grep -v "time\.Second\|time\.Minute\|time\.Hour" || true)
if [ -n "$TIMING_LITERALS" ]; then
    echo "⚠️  WARNING: Found timing literals in E2E tests:"
    echo "$TIMING_LITERALS"
    echo "Consider using fixtures.LoadTestConfig().Timing.* for timing values"
fi

# Check for error code literals (should use fixtures)
echo "Checking for error code literals..."
ERROR_LITERALS=$(grep -r "\"BUSY\"\|\"INVALID_RANGE\"\|\"UNAVAILABLE\"\|\"INTERNAL\"" test/e2e/ || true)
if [ -n "$ERROR_LITERALS" ]; then
    echo "⚠️  WARNING: Found error code literals in E2E tests:"
    echo "$ERROR_LITERALS"
    echo "Consider using fixtures.BusyError(), fixtures.RangeError(), etc."
fi

# Check for harness access only (ARCH-REMEDY-04)
echo "Checking for harness access restrictions..."
HARNESS_ACCESS=$(grep -r "server\." test/e2e/ | grep -vE "server\.(URL|Shutdown)" || true)
if [ -n "$HARNESS_ACCESS" ]; then
    echo "❌ FAIL: Found forbidden harness access in E2E tests:"
    echo "$HARNESS_ACCESS"
    echo "Only server.URL and server.Shutdown are allowed"
    exit 1
fi

# Check for no wrappers (ARCH-REMEDY-06) in Go code only
echo "Checking for wrapper elimination..."
WRAPPERS=$(grep -r --include="*.go" -E "radioManagerWrapper|radioManagerAdapter" internal test/e2e || true)
if [ -n "$WRAPPERS" ]; then
    echo "❌ FAIL: Found wrapper usage in code:"
    echo "$WRAPPERS"
    echo "Wrappers should be eliminated after port implementation"
    exit 1
fi

# Check for API ports only (ARCH-REMEDY-03) - non-E2E scope
if [ "$ANTI_PEEK_SCOPE" != "e2e" ]; then
    echo "Checking for API server port usage..."
    API_CONCRETE=$(grep -r "\*command\.Orchestrator\|\*telemetry\.Hub\|\*radio\.Manager" internal/api/server.go || true)
    if [ -n "$API_CONCRETE" ]; then
        echo "❌ FAIL: Found concrete types in API server:"
        echo "$API_CONCRETE"
        echo "API server should use ports only"
        exit 1
    fi
fi

# Check for no interface{} in api/command (ARCH-REMEDY-05) - non-E2E scope
if [ "$ANTI_PEEK_SCOPE" != "e2e" ]; then
    echo "Checking for interface{} elimination..."
    INTERFACE_ANY=$(grep -r "interface{}" internal/api internal/command || true)
    if [ -n "$INTERFACE_ANY" ]; then
        echo "❌ FAIL: Found interface{} usage in api/command:"
        echo "$INTERFACE_ANY"
        echo "Use concrete types or DTOs instead of interface{}"
        exit 1
    fi
fi

# Check for no internal imports in e2e (ARCH-REMEDY-04)
echo "Checking for internal imports in E2E..."
E2E_INTERNAL=$(grep -r "github.com/.*/internal/" test/e2e/ --include="*.go" || true)
if [ -n "$E2E_INTERNAL" ]; then
    echo "❌ FAIL: Found internal imports in E2E tests:"
    echo "$E2E_INTERNAL"
    echo "E2E tests should not import internal packages"
    exit 1
fi

echo "✅ PASS: All anti-peek checks passed"
echo "E2E tests are properly black-box and spec-driven"
echo "✅ E2E tests enforce API-as-ground-truth principle"
echo "✅ Architectural refactor rules enforced"
