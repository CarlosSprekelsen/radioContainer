#!/bin/bash
# Cross-Doc Compliance Check: Timing Literals
# Source: Cross-Doc-Consistency-Matrix-v1 §3.3
# Ensures timing literals only appear in internal/config/*

set -e

echo "🔍 Checking timing literal compliance..."

# Check for hardcoded timing literals outside config
echo "  📋 Checking for hardcoded timing literals outside config..."
TIMING_VIOLATIONS=$(find . -name "*.go" -not -path "./internal/config/*" -not -path "./tools/*" -not -path "./cmd/*" -not -path "./internal/api/server.go" -exec grep -l "time\.Second\|time\.Minute\|time\.Hour" {} \; || true)

if [ -n "$TIMING_VIOLATIONS" ]; then
    echo "  ❌ Found timing literals outside config:"
    for file in $TIMING_VIOLATIONS; do
        echo "    - $file"
        grep -n "time\.Second\|time\.Minute\|time\.Hour" "$file" | head -3
    done
    echo "  💡 All timing values should be loaded from CB-TIMING v0.3 via config"
    exit 1
fi
echo "  ✅ No timing literals found outside config"

# Check that config package references CB-TIMING
echo "  📋 Checking config package CB-TIMING references..."
if ! grep -q "CB-TIMING v0.3" internal/config/*.go; then
    echo "  ❌ CB-TIMING v0.3 reference not found in config"
    exit 1
fi
echo "  ✅ CB-TIMING v0.3 reference found in config"

# Check that timing values are loaded from baseline
echo "  📋 Checking timing baseline loading..."
if ! grep -q "LoadCBTimingBaseline" internal/config/*.go; then
    echo "  ❌ LoadCBTimingBaseline function not found"
    exit 1
fi
echo "  ✅ LoadCBTimingBaseline function found"

# Check for specific CB-TIMING parameters
echo "  📋 Checking CB-TIMING parameter coverage..."
CB_TIMING_PARAMS=("HeartbeatInterval" "HeartbeatJitter" "HeartbeatTimeout" "ProbeNormalInterval" "ProbeRecoveringInitial" "ProbeRecoveringBackoff" "ProbeRecoveringMax" "ProbeOfflineInitial" "ProbeOfflineBackoff" "ProbeOfflineMax" "CommandTimeoutSetPower" "CommandTimeoutSetChannel" "CommandTimeoutSelectRadio" "CommandTimeoutGetState" "EventBufferSize" "EventBufferRetention")

for param in "${CB_TIMING_PARAMS[@]}"; do
    if ! grep -q "$param" internal/config/*.go; then
        echo "  ❌ CB-TIMING parameter $param not found in config"
        exit 1
    fi
done
echo "  ✅ All CB-TIMING parameters found in config"

# Check that other packages reference config for timing
echo "  📋 Checking other packages use config for timing..."
PACKAGES_WITH_TIMING=("internal/telemetry" "internal/audit" "internal/radio" "internal/api")

for pkg in "${PACKAGES_WITH_TIMING[@]}"; do
    if [ -d "$pkg" ]; then
        if grep -q "time\." "$pkg"/*.go 2>/dev/null; then
            if ! grep -q "config\." "$pkg"/*.go 2>/dev/null; then
                echo "  ⚠️  Package $pkg uses time but may not use config"
            fi
        fi
    fi
done
echo "  ✅ Timing usage in packages checked"

# Check for environment variable overrides
echo "  📋 Checking environment variable override support..."
if ! grep -q "RCC_TIMING_" internal/config/*.go; then
    echo "  ❌ Environment variable override support not found"
    exit 1
fi
echo "  ✅ Environment variable override support found"

# Check for validation rules
echo "  📋 Checking timing validation rules..."
if ! grep -q "ValidateTiming" internal/config/*.go; then
    echo "  ❌ Timing validation not found"
    exit 1
fi
echo "  ✅ Timing validation found"

# Check that main.go loads config
echo "  📋 Checking main.go loads config..."
if ! grep -q "config\.Load" cmd/rcc/main.go; then
    echo "  ❌ Config loading not found in main.go"
    exit 1
fi
echo "  ✅ Config loading found in main.go"

echo "✅ Timing literal compliance check passed"
