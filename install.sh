#!/bin/bash
# install.sh
# One-command installer for PS-IDE-Go
# Usage: curl -sSL https://raw.githubusercontent.com/LaurieRhodes/ps-ide-go/main/install.sh | bash

set -e

REPO="LaurieRhodes/ps-ide-go"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
GITHUB_API="https://api.github.com/repos/$REPO/releases/latest"

GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}   PS-IDE-Go Installer${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check OS
if [[ "$OSTYPE" != "linux-gnu"* ]]; then
    echo -e "${RED}Error: This installer only supports Linux${NC}"
    echo "For Windows/macOS, download from: https://github.com/$REPO/releases"
    exit 1
fi

# Check architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        BINARY_NAME="ps-ide-linux-amd64.tar.gz"
        ;;
    aarch64|arm64)
        BINARY_NAME="ps-ide-linux-arm64.tar.gz"
        ;;
    *)
        echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

echo "Detected: Linux $ARCH"
echo ""

# Install runtime dependencies
echo -e "${GREEN}[1/4] Installing runtime dependencies...${NC}"
if command -v apt-get &> /dev/null; then
    echo "Using apt package manager..."
    sudo apt-get update -qq
    
    # PowerShell
    if ! command -v pwsh &> /dev/null; then
        echo "  Installing PowerShell..."
        sudo apt-get install -y -qq powershell > /dev/null
    else
        echo "  ✅ PowerShell already installed"
    fi
    
    # GTK3 runtime
    echo "  Installing GTK3 runtime..."
    sudo apt-get install -y -qq libgtk-3-0 > /dev/null
    
elif command -v dnf &> /dev/null; then
    echo "Using dnf package manager..."
    sudo dnf install -y -q powershell gtk3 > /dev/null
elif command -v pacman &> /dev/null; then
    echo "Using pacman package manager..."
    sudo pacman -S --noconfirm --needed powershell gtk3 > /dev/null
else
    echo -e "${RED}Error: No supported package manager found${NC}"
    echo "Please install: powershell, libgtk-3-0"
    exit 1
fi

echo ""

# Get latest version
echo -e "${GREEN}[2/4] Fetching latest release...${NC}"
if ! VERSION=$(curl -s "$GITHUB_API" | grep -oP '"tag_name":\s*"\K[^"]+'); then
    echo -e "${RED}Error: Could not fetch latest version${NC}"
    exit 1
fi

echo "  Latest version: $VERSION"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/$BINARY_NAME"
echo ""

# Download binary
echo -e "${GREEN}[3/4] Downloading ps-ide...${NC}"
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

if ! curl -sL "$DOWNLOAD_URL" -o "$BINARY_NAME"; then
    echo -e "${RED}Error: Download failed${NC}"
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Extract and install
tar xzf "$BINARY_NAME"

echo ""
echo -e "${GREEN}[4/4] Installing binary...${NC}"

# Create install directory if it doesn't exist
mkdir -p "$INSTALL_DIR"

# Move binary
mv ps-ide "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/ps-ide"

# Clean up
cd - > /dev/null
rm -rf "$TEMP_DIR"

echo "  Installed to: $INSTALL_DIR/ps-ide"
echo ""

# Check if install dir is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo -e "${BLUE}Note: $INSTALL_DIR is not in your PATH${NC}"
    echo "Add this line to your ~/.bashrc or ~/.zshrc:"
    echo ""
    echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
    echo ""
    echo "Then run: source ~/.bashrc"
    echo ""
fi

# Success message
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}✅ Installation complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Run ps-ide with:"
if [[ ":$PATH:" == *":$INSTALL_DIR:"* ]]; then
    echo "  ps-ide"
else
    echo "  $INSTALL_DIR/ps-ide"
fi
echo ""
echo "Or open a PowerShell file:"
if [[ ":$PATH:" == *":$INSTALL_DIR:"* ]]; then
    echo "  ps-ide script.ps1"
else
    echo "  $INSTALL_DIR/ps-ide script.ps1"
fi
echo ""
