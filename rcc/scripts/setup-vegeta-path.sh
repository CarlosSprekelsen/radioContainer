#!/bin/bash

# Setup Vegeta PATH for Radio Control Container
# This script ensures Vegeta is available in PATH for CI/CD and local development

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔧 Setting up Vegeta PATH${NC}"
echo "=========================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed. Please install Go first.${NC}"
    exit 1
fi

# Get Go bin directory
GO_BIN_DIR=$(go env GOPATH)/bin
echo -e "${YELLOW}📁 Go bin directory: $GO_BIN_DIR${NC}"

# Check if Vegeta exists
if [ -f "$GO_BIN_DIR/vegeta" ]; then
    echo -e "${GREEN}✅ Vegeta found at: $GO_BIN_DIR/vegeta${NC}"
else
    echo -e "${RED}❌ Vegeta not found at: $GO_BIN_DIR/vegeta${NC}"
    echo -e "${YELLOW}💡 Install with: go install github.com/tsenart/vegeta@latest${NC}"
    exit 1
fi

# Add to PATH for current session
export PATH=$PATH:$GO_BIN_DIR
echo -e "${GREEN}✅ Added $GO_BIN_DIR to PATH${NC}"

# Test Vegeta
if command -v vegeta &> /dev/null; then
    echo -e "${GREEN}✅ Vegeta is now available in PATH${NC}"
    vegeta -version
else
    echo -e "${RED}❌ Vegeta still not available in PATH${NC}"
    exit 1
fi

# Create environment setup script
ENV_SCRIPT="scripts/vegeta-env.sh"
cat > "$ENV_SCRIPT" << EOF
#!/bin/bash
# Vegeta environment setup
export PATH=\$PATH:$GO_BIN_DIR
EOF

chmod +x "$ENV_SCRIPT"
echo -e "${GREEN}✅ Created environment script: $ENV_SCRIPT${NC}"

# Update Makefile to use environment script
MAKEFILE_TARGET="vegeta-env: ## Setup Vegeta environment
	@export PATH=\$$PATH:$GO_BIN_DIR"

echo -e "${BLUE}📋 Usage:${NC}"
echo "   # Source environment for current session"
echo "   source $ENV_SCRIPT"
echo ""
echo "   # Or add to your shell profile"
echo "   echo 'export PATH=\$PATH:$GO_BIN_DIR' >> ~/.bashrc"
echo ""
echo "   # Run Vegeta scenarios"
echo "   bash test/perf/vegeta_scenarios.sh"
echo ""
echo -e "${GREEN}🎉 Vegeta PATH setup complete!${NC}"
