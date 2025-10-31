#!/bin/bash
# Next Steps - Run this script to initialize and build the project

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}PS-IDE-Go - Next Steps Script${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Step 1: Initialize Go module
echo -e "${GREEN}Step 1: Initializing Go module...${NC}"
if [ -f "go.sum" ]; then
    echo "Go module already initialized"
else
    go mod init github.com/laurie/ps-ide-go
    echo "✓ Go module initialized"
fi
echo ""

# Step 2: Download dependencies
echo -e "${GREEN}Step 2: Downloading dependencies...${NC}"
echo "This may take a minute..."
go mod download
go mod tidy
echo "✓ Dependencies downloaded"
echo ""

# Step 3: Verify PowerShell
echo -e "${GREEN}Step 3: Verifying PowerShell installation...${NC}"
if command -v pwsh &> /dev/null; then
    echo "✓ PowerShell found: $(pwsh --version)"
else
    echo -e "${YELLOW}⚠ PowerShell not found!${NC}"
    echo "Install with: sudo apt install powershell"
    exit 1
fi
echo ""

# Step 4: Check GUI dependencies
echo -e "${GREEN}Step 4: Checking GUI dependencies...${NC}"
if command -v gcc &> /dev/null; then
    echo "✓ GCC found: $(gcc --version | head -n1)"
else
    echo -e "${YELLOW}⚠ GCC not found! Run: sudo apt install gcc${NC}"
fi

if pkg-config --exists gl; then
    echo "✓ OpenGL libraries found"
else
    echo -e "${YELLOW}⚠ OpenGL libraries not found! Run: sudo apt install libgl1-mesa-dev${NC}"
fi

if pkg-config --exists x11; then
    echo "✓ X11 libraries found"
else
    echo -e "${YELLOW}⚠ X11 libraries not found! Run: sudo apt install xorg-dev${NC}"
fi
echo ""

# Step 5: Build the application
echo -e "${GREEN}Step 5: Building the application...${NC}"
go build -o ps-ide ./cmd/ps-ide
if [ $? -eq 0 ]; then
    echo "✓ Build successful!"
else
    echo -e "${YELLOW}✗ Build failed. Check errors above.${NC}"
    exit 1
fi
echo ""

# Step 6: Test run
echo -e "${GREEN}Step 6: Testing the application...${NC}"
echo "Checking if the binary works..."
if [ -f "./ps-ide" ]; then
    echo "✓ Binary created: ./ps-ide"
    ls -lh ./ps-ide
else
    echo -e "${YELLOW}✗ Binary not found${NC}"
    exit 1
fi
echo ""

# Success!
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}✓ Setup Complete!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "You can now run the application with:"
echo -e "${BLUE}  ./ps-ide${NC}"
echo ""
echo "Or build using make:"
echo -e "${BLUE}  make build${NC}"
echo -e "${BLUE}  make run${NC}"
echo ""
echo "Check the documentation:"
echo "  - README.md - Project overview"
echo "  - docs/QUICKSTART.md - User guide"
echo "  - docs/DEVELOPMENT.md - Developer guide"
echo "  - PROJECT_SUMMARY.md - Complete overview"
echo ""
echo -e "${GREEN}Happy coding! 🚀${NC}"
