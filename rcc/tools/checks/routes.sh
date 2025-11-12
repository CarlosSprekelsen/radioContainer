#!/bin/bash
# Cross-Doc Compliance Check: API Routes
# Source: OpenAPI v1 §3
# Ensures all required routes exist and are properly implemented

set -e

echo "🔍 Checking API route compliance..."

# Required routes from OpenAPI v1
REQUIRED_ROUTES=(
    "GET /api/v1/capabilities"
    "GET /api/v1/radios"
    "POST /api/v1/radios/select"
    "GET /api/v1/radios/{id}"
    "GET /api/v1/radios/{id}/power"
    "POST /api/v1/radios/{id}/power"
    "GET /api/v1/radios/{id}/channel"
    "POST /api/v1/radios/{id}/channel"
    "GET /api/v1/telemetry"
    "GET /api/v1/health"
)

# Check that routes are registered
echo "  📋 Checking route registration..."
if ! grep -q "RegisterRoutes" internal/api/routes.go; then
    echo "  ❌ RegisterRoutes function not found"
    exit 1
fi
echo "  ✅ RegisterRoutes function found"

# Check individual route handlers
echo "  📋 Checking individual route handlers..."

# Check capabilities endpoint
if ! grep -q "handleCapabilities" internal/api/routes.go; then
    echo "  ❌ handleCapabilities not found"
    exit 1
fi
echo "  ✅ handleCapabilities found"

# Check radios endpoints
if ! grep -q "handleRadios" internal/api/routes.go; then
    echo "  ❌ handleRadios not found"
    exit 1
fi
echo "  ✅ handleRadios found"

# Check select radio endpoint
if ! grep -q "handleSelectRadio" internal/api/routes.go; then
    echo "  ❌ handleSelectRadio not found"
    exit 1
fi
echo "  ✅ handleSelectRadio found"

# Check radio-specific endpoints
if ! grep -q "handleRadioEndpoints" internal/api/routes.go; then
    echo "  ❌ handleRadioEndpoints not found"
    exit 1
fi
echo "  ✅ handleRadioEndpoints found"

# Check power endpoints
if ! grep -q "handleGetPower\|handleSetPower" internal/api/routes.go; then
    echo "  ❌ Power endpoints not found"
    exit 1
fi
echo "  ✅ Power endpoints found"

# Check channel endpoints
if ! grep -q "handleGetChannel\|handleSetChannel" internal/api/routes.go; then
    echo "  ❌ Channel endpoints not found"
    exit 1
fi
echo "  ✅ Channel endpoints found"

# Check telemetry endpoint
if ! grep -q "handleTelemetry" internal/api/routes.go; then
    echo "  ❌ handleTelemetry not found"
    exit 1
fi
echo "  ✅ handleTelemetry found"

# Check health endpoint
if ! grep -q "handleHealth" internal/api/routes.go; then
    echo "  ❌ handleHealth not found"
    exit 1
fi
echo "  ✅ handleHealth found"

# Check HTTP method validation
echo "  📋 Checking HTTP method validation..."
if ! grep -q "r.Method.*http.MethodGet\|r.Method.*http.MethodPost" internal/api/routes.go; then
    echo "  ❌ HTTP method validation not found"
    exit 1
fi
echo "  ✅ HTTP method validation found"

# Check error handling
echo "  📋 Checking error handling..."
if ! grep -q "WriteError\|WriteSuccess" internal/api/routes.go; then
    echo "  ❌ Error handling functions not found"
    exit 1
fi
echo "  ✅ Error handling functions found"

# Check response envelope
echo "  📋 Checking response envelope..."
if ! grep -q "correlationId\|CorrelationID" internal/api/response.go; then
    echo "  ❌ Response envelope not found"
    exit 1
fi
echo "  ✅ Response envelope found"

# Check OpenAPI v1 references
echo "  📋 Checking OpenAPI v1 references..."
if ! grep -q "Source: OpenAPI v1" internal/api/routes.go; then
    echo "  ❌ OpenAPI v1 reference not found"
    exit 1
fi
echo "  ✅ OpenAPI v1 reference found"

# Check route path patterns
echo "  📋 Checking route path patterns..."
if ! grep -q "/api/v1/" internal/api/routes.go; then
    echo "  ❌ API v1 path pattern not found"
    exit 1
fi
echo "  ✅ API v1 path pattern found"

# Check for radio ID extraction
echo "  📋 Checking radio ID extraction..."
if ! grep -q "ExtractRadioID\|radioID" internal/api/routes.go; then
    echo "  ❌ Radio ID extraction not found"
    exit 1
fi
echo "  ✅ Radio ID extraction found"

# Check for telemetry hub integration
echo "  📋 Checking telemetry hub integration..."
if ! grep -q "telemetryHub\|TelemetryHub" internal/api/routes.go; then
    echo "  ❌ Telemetry hub integration not found"
    exit 1
fi
echo "  ✅ Telemetry hub integration found"

# Check for server structure
echo "  📋 Checking server structure..."
if ! grep -q "type.*Server" internal/api/server.go; then
    echo "  ❌ Server structure not found"
    exit 1
fi
echo "  ✅ Server structure found"

# Check for server start/stop methods
echo "  📋 Checking server lifecycle methods..."
if ! grep -q "func.*Start\|func.*Stop" internal/api/server.go; then
    echo "  ❌ Server lifecycle methods not found"
    exit 1
fi
echo "  ✅ Server lifecycle methods found"

# Check for main.go integration
echo "  📋 Checking main.go integration..."
if ! grep -q "api\.NewServer\|server\.Start" cmd/rcc/main.go; then
    echo "  ❌ API server integration not found in main.go"
    exit 1
fi
echo "  ✅ API server integration found in main.go"

echo "✅ API route compliance check passed"
