#!/bin/bash
# install-runtime.sh
# Minimal runtime dependencies for ps-ide binary users

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}PS-IDE-Go Runtime Installer${NC}"
echo "Installing minimal runtime dependencies..."
echo ""

# Update package list
echo "Updating package list..."
sudo apt update > /dev/null

# Check and install PowerShell
if ! command -v pwsh &> /dev/null; then
    echo -e "${GREEN}Installing PowerShell...${NC}"
    sudo apt install -y powershell
else
    echo "✅ PowerShell already installed: $(pwsh --version)"
fi

# Check and install GTK3 runtime
echo -e "${GREEN}Installing GTK3 runtime libraries...${NC}"
sudo apt install -y libgtk-3-0

echo ""
echo -e "${GREEN}✅ Runtime dependencies installed!${NC}"
echo ""
echo "Next steps:"
echo "1. Download the latest binary:"
echo "   https://github.com/LaurieRhodes/ps-ide-go/releases"
echo ""
echo "2. Extract and run:"
echo "   tar xzf ps-ide-linux-amd64.tar.gz"
echo "   ./ps-ide"
echo ""
