#!/bin/bash
# install-dev-deps.sh
# Development dependencies for building PS-IDE-Go from source

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Installing PS-IDE-Go development dependencies...${NC}"
echo ""

# Update package list
echo "Updating package list..."
sudo apt update

# Check and install Go
if ! command -v go &> /dev/null; then
    echo -e "${YELLOW}Installing Go...${NC}"
    sudo apt install -y golang-go
else
    echo "✅ Go already installed: $(go version)"
fi

# Check and install PowerShell
if ! command -v pwsh &> /dev/null; then
    echo -e "${YELLOW}Installing PowerShell...${NC}"
    sudo apt install -y powershell
else
    echo "✅ PowerShell already installed: $(pwsh --version)"
fi

# Install GTK3 development dependencies
echo -e "${YELLOW}Installing GTK3 development libraries...${NC}"
sudo apt install -y build-essential pkg-config \
    libgtk-3-dev libglib2.0-dev libcairo2-dev libpango1.0-dev

echo ""
echo -e "${GREEN}✅ Development environment ready!${NC}"
echo ""
echo "Next steps:"
echo "1. Download Go module dependencies:"
echo "   go mod download"
echo ""
echo "2. Build the application:"
echo "   make build"
echo "   # or: go build -o ps-ide ./cmd/ps-ide"
echo ""
echo "3. Run the application:"
echo "   ./ps-ide"
echo ""
