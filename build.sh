#!/bin/bash

# Build script for PS-IDE-Go

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Building PS-IDE-Go...${NC}"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

# Check if PowerShell is installed
if ! command -v pwsh &> /dev/null; then
    echo -e "${YELLOW}Warning: PowerShell (pwsh) is not installed${NC}"
    echo -e "${YELLOW}The application will not work without PowerShell${NC}"
fi

# Clean previous builds
echo "Cleaning previous builds..."
rm -f ps-ide

# Get version from git or use default
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')

echo "Version: $VERSION"
echo "Build time: $BUILD_TIME"

# Download dependencies
echo "Downloading dependencies..."
go mod download

# Build the application
echo "Compiling..."
go build \
    -ldflags="-X 'main.Version=$VERSION' -X 'main.BuildTime=$BUILD_TIME'" \
    -o ps-ide \
    ./cmd/ps-ide

if [ $? -eq 0 ]; then
    echo -e "${GREEN}Build successful!${NC}"
    echo -e "Binary: ${GREEN}./ps-ide${NC}"
    echo ""
    echo "Run with: ./ps-ide"
else
    echo -e "${RED}Build failed!${NC}"
    exit 1
fi
