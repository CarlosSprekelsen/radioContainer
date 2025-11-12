#!/bin/bash
# Architecture Guardrails - Enforce error handling and API patterns
# Source: Architecture §8.5, §8.5.1 Error normalization and diagnostics

set -e

echo "🔍 Running architecture guardrails..."

# Check 1: No raw fmt.Errorf in public methods (exclude test/mock adapters)
echo "Checking for raw fmt.Errorf in public methods..."
RAW_ERRORS=$(grep -r "return fmt\.Errorf(" internal/command/ internal/adapter/ --include="*.go" | grep -v "//" | grep -v "test" | grep -v "fake" | grep -v "mock" || true)
if [ -n "$RAW_ERRORS" ]; then
    echo "❌ FAIL: Found raw fmt.Errorf in public methods:"
    echo "$RAW_ERRORS"
    echo "Use adapter.NewVendorError() or normalized adapter errors instead"
    exit 1
fi
echo "✅ PASS: No raw fmt.Errorf in public methods"

# Check 2: ToAPIError must unwrap VendorError
echo "Checking ToAPIError VendorError unwrapping..."
if ! grep -q "errors\.As.*vendorErr" internal/api/errors.go; then
    echo "❌ FAIL: ToAPIError missing errors.As(err, &vendorErr) path"
    echo "Add VendorError unwrapping in ToAPIError function"
    exit 1
fi
echo "✅ PASS: ToAPIError has VendorError unwrapping"

# Check 3: No http.Error usage in API layer
echo "Checking for http.Error usage in API layer..."
HTTP_ERRORS=$(grep -r "http\.Error" internal/api/ --include="*.go" || true)
if [ -n "$HTTP_ERRORS" ]; then
    echo "❌ FAIL: Found http.Error usage in API layer:"
    echo "$HTTP_ERRORS"
    echo "Use WriteError() or ToAPIError() instead"
    exit 1
fi
echo "✅ PASS: No http.Error usage in API layer"

# Check 4: No direct WriteHeader outside ToAPIError paths
echo "Checking for direct WriteHeader usage..."
DIRECT_WRITES=$(grep -r "w\.WriteHeader(" internal/api/ --include="*.go" | grep -v "ToAPIError" | grep -v "WriteError" | grep -v "response.go" || true)
if [ -n "$DIRECT_WRITES" ]; then
    echo "❌ FAIL: Found direct WriteHeader usage outside ToAPIError paths:"
    echo "$DIRECT_WRITES"
    echo "Use ToAPIError() or WriteError() for unified error envelope"
    exit 1
fi
echo "✅ PASS: No direct WriteHeader usage"

# Check 5: Timing literals in non-test code
echo "Checking for timing literals in non-test code..."
TIMING_LITERALS=$(grep -r "time\." internal/ --include="*.go" | grep -v "test" | grep -v "time\.Since\|time\.Now\|time\.Time" || true)
if [ -n "$TIMING_LITERALS" ]; then
    echo "⚠️  WARNING: Found timing literals in non-test code:"
    echo "$TIMING_LITERALS"
    echo "Consider using config.Timing.* values from CB-TIMING"
fi

echo "✅ PASS: All architecture guardrails passed"
echo "✅ Error handling architecture compliance verified"
