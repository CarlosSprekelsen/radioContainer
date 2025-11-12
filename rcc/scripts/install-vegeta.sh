#!/bin/bash

# Install Vegeta for HTTP load testing
# This is simpler and more appropriate than k6 for this Go project

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Installing Vegeta for HTTP Load Testing${NC}"
echo "=============================================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed. Please install Go first.${NC}"
    exit 1
fi

# Check if vegeta is already installed
if command -v vegeta &> /dev/null; then
    echo -e "${YELLOW}⚠️  Vegeta is already installed${NC}"
    vegeta -version
    echo -e "${GREEN}✅ Vegeta is ready to use${NC}"
    exit 0
fi

# Check if vegeta is installed but not in PATH
if [ -f "/home/$USER/go/bin/vegeta" ]; then
    echo -e "${YELLOW}⚠️  Vegeta is installed but not in PATH${NC}"
    echo -e "${BLUE}🔧 Adding Go bin directory to PATH...${NC}"
    export PATH=$PATH:/home/$USER/go/bin
    if command -v vegeta &> /dev/null; then
        echo -e "${GREEN}✅ Vegeta is now available in PATH${NC}"
        vegeta -version
        echo -e "${GREEN}✅ Vegeta is ready to use${NC}"
        exit 0
    fi
fi

echo -e "${YELLOW}📦 Installing Vegeta...${NC}"

# Install vegeta
go install github.com/tsenart/vegeta@latest

# Verify installation
if command -v vegeta &> /dev/null; then
    echo -e "${GREEN}✅ Vegeta installed successfully${NC}"
    vegeta version
else
    echo -e "${RED}❌ Vegeta installation failed${NC}"
    exit 1
fi

echo -e "${GREEN}🎉 Vegeta installation complete!${NC}"
echo ""
echo -e "${BLUE}📋 Usage:${NC}"
echo "   # Run load tests"
echo "   bash test/perf/vegeta_scenarios.sh"
echo ""
echo "   # Manual load testing"
echo "   echo 'GET http://localhost:8080/api/v1/radios' | vegeta attack -duration=30s -rate=100 | vegeta report"
echo ""
echo -e "${BLUE}🔧 Integration:${NC}"
echo "   - Vegeta scenarios: test/perf/vegeta_scenarios.sh"
echo "   - Performance monitoring: ./scripts/performance-monitor.sh"
echo "   - CI/CD integration: .github/workflows/performance-benchmarks.yml"
