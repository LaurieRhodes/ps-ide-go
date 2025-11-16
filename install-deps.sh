#!/bin/bash

# Install script for PS-IDE-Go

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Installing PS-IDE-Go dependencies...${NC}"

# Update package list
echo "Updating package list..."
sudo apt update

# Check and install Go
if ! command -v go &> /dev/null; then
    echo "Installing Go..."
    sudo apt install -y golang-go
else
    echo "Go is already installed: $(go version)"
fi

# Check and install PowerShell
if ! command -v pwsh &> /dev/null; then
    echo "Installing PowerShell..."
    sudo apt install -y powershell
else
    echo "PowerShell is already installed: $(pwsh --version)"
fi

# Install GUI development dependencies
echo "Installing GUI dependencies..."
sudo apt install -y build-essential pkg-config \
    libgtk-3-dev libglib2.0-dev libcairo2-dev \
    libpango1.0-dev gcc libgl1-mesa-dev xorg-dev

# Install Go dependencies
echo "Installing Go module dependencies..."
cd "$(dirname "$0")"
go mod download

echo -e "${GREEN}Installation complete!${NC}"
echo ""
echo "Next steps:"
echo "1. Run: go build -o ps-ide ./cmd/ps-ide"
echo "2. Run: ./ps-ide"
