#!/bin/bash
set -e

echo "🔍 Pre-publish linting for ps-ide-go..."
echo

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

cd "$(dirname "$0")"

# 1. Format
echo "📝 Formatting code..."
go fmt ./...
echo -e "${GREEN}✓ Formatting complete${NC}\n"

# 2. Go vet
echo "🔎 Running go vet..."
if go vet ./... 2>&1 | tee /tmp/vet.out; then
    echo -e "${GREEN}✓ go vet passed${NC}\n"
else
    echo -e "${RED}✗ go vet found issues (see above)${NC}\n"
    exit 1
fi

# 3. Staticcheck (if available)
if command -v staticcheck &> /dev/null; then
    echo "🔍 Running staticcheck..."
    if staticcheck ./...; then
        echo -e "${GREEN}✓ staticcheck passed${NC}\n"
    else
        echo -e "${YELLOW}⚠ staticcheck found issues${NC}\n"
    fi
else
    echo -e "${YELLOW}⚠ staticcheck not installed (optional)${NC}\n"
fi

# 4. Run tests
echo "🧪 Running tests..."
if go test -race -coverprofile=coverage.out ./...; then
    echo -e "${GREEN}✓ Tests passed${NC}\n"
    
    # Show coverage
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    echo "📊 Coverage: $COVERAGE"
else
    echo -e "${RED}✗ Tests failed${NC}\n"
    exit 1
fi

# 5. golangci-lint (if available)
GOLANGCI_LINT=""
if command -v golangci-lint &> /dev/null; then
    GOLANGCI_LINT="golangci-lint"
elif [ -x "$HOME/go/bin/golangci-lint" ]; then
    GOLANGCI_LINT="$HOME/go/bin/golangci-lint"
fi

if [ -n "$GOLANGCI_LINT" ]; then
    echo "🔧 Running golangci-lint..."
    if $GOLANGCI_LINT run --timeout=5m; then
        echo -e "${GREEN}✓ golangci-lint passed${NC}\n"
    else
        echo -e "${RED}✗ golangci-lint found issues${NC}\n"
        echo "Run: $GOLANGCI_LINT run --fix"
        exit 1
    fi
else
    echo -e "${YELLOW}⚠ golangci-lint not installed${NC}"
    echo "Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin"
    echo
fi

# 6. Build check
echo "🏗️  Building..."
if go build -v ./cmd/ps-ide; then
    echo -e "${GREEN}✓ Build successful${NC}\n"
else
    echo -e "${RED}✗ Build failed${NC}\n"
    exit 1
fi

# 7. Verify dependencies
echo "📦 Verifying dependencies..."
go mod tidy
go mod verify
echo -e "${GREEN}✓ Dependencies verified${NC}\n"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}✓ All checks passed! Ready to publish${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
