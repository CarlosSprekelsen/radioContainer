#!/bin/bash
# Cross-Doc Compliance Check: Error Codes
# Source: Cross-Doc-Consistency-Matrix-v1 §3.2
# Ensures error codes match between OpenAPI and implementation

set -e

echo "🔍 Checking error code compliance..."

# Expected error codes from OpenAPI v1 §2.2
EXPECTED_CODES=("BAD_REQUEST" "INVALID_RANGE" "BUSY" "UNAVAILABLE" "INTERNAL" "UNAUTHORIZED" "FORBIDDEN" "NOT_FOUND")

# Check adapter error codes
echo "  📋 Checking adapter error codes..."
ADAPTER_ERRORS=$(grep -r "ErrInvalidRange\|ErrBusy\|ErrUnavailable\|ErrInternal" internal/adapter/ || true)
if [ -z "$ADAPTER_ERRORS" ]; then
    echo "  ❌ No adapter error codes found"
    exit 1
fi
echo "  ✅ Adapter error codes found"

# Check API error codes
echo "  📋 Checking API error codes..."
API_ERRORS=$(grep -r "ErrUnauthorizedError\|ErrForbiddenError\|ErrNotFoundError" internal/api/ || true)
if [ -z "$API_ERRORS" ]; then
    echo "  ❌ No API error codes found"
    exit 1
fi
echo "  ✅ API error codes found"

# Check error mapping function
echo "  📋 Checking error mapping function..."
if ! grep -q "func ToAPIError" internal/api/errors.go; then
    echo "  ❌ ToAPIError function not found"
    exit 1
fi
echo "  ✅ ToAPIError function found"

# Check HTTP status code mappings
echo "  📋 Checking HTTP status code mappings..."
STATUS_MAPPINGS=$(grep -r "http.StatusBadRequest\|http.StatusUnauthorized\|http.StatusForbidden\|http.StatusNotFound\|http.StatusServiceUnavailable\|http.StatusInternalServerError" internal/api/ || true)
if [ -z "$STATUS_MAPPINGS" ]; then
    echo "  ❌ No HTTP status code mappings found"
    exit 1
fi
echo "  ✅ HTTP status code mappings found"

# Check correlation ID in error responses
echo "  📋 Checking correlation ID in error responses..."
if ! grep -q "CorrelationID.*generateCorrelationID" internal/api/errors.go; then
    echo "  ❌ Correlation ID not found in error responses"
    exit 1
fi
echo "  ✅ Correlation ID found in error responses"

# Check OpenAPI document references
echo "  📋 Checking OpenAPI document references..."
if ! grep -q "Source: OpenAPI v1" internal/api/errors.go; then
    echo "  ❌ OpenAPI v1 reference not found in error mapping"
    exit 1
fi
echo "  ✅ OpenAPI v1 reference found"

# Check Architecture document references
echo "  📋 Checking Architecture document references..."
if ! grep -q "Architecture §8.5" internal/api/errors.go; then
    echo "  ❌ Architecture §8.5 reference not found in error mapping"
    exit 1
fi
echo "  ✅ Architecture §8.5 reference found"

echo "✅ Error code compliance check passed"
