#!/bin/bash

set -e

echo "====================================="
echo "Marimo ERP Dependency Update Script"
echo "====================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Go dependencies
echo "1. Updating Go dependencies..."
echo "================================"

if command -v go &> /dev/null; then
    cd /home/user/marimo

    # List outdated modules
    echo "Checking for outdated Go modules..."
    go list -u -m all | grep '\['

    # Update all dependencies
    echo ""
    echo "Updating Go modules..."
    go get -u ./...

    # Tidy up
    go mod tidy

    # Verify
    go mod verify

    print_success "Go dependencies updated"
else
    print_error "Go not found, skipping Go dependencies"
fi

echo ""
echo "2. Updating Frontend dependencies..."
echo "====================================="

if command -v npm &> /dev/null; then
    cd /home/user/marimo/frontend

    # Check for outdated packages
    echo "Checking for outdated npm packages..."
    npm outdated || true

    echo ""
    echo "Updating npm packages..."
    npm update

    # Run npm audit
    echo ""
    echo "Running security audit..."
    npm audit || print_warning "Some vulnerabilities found, run 'npm audit fix' manually"

    print_success "Frontend dependencies updated"
else
    print_error "npm not found, skipping frontend dependencies"
fi

echo ""
echo "3. Updating Mobile dependencies..."
echo "===================================="

if [ -d "/home/user/marimo/mobile" ]; then
    cd /home/user/marimo/mobile

    if command -v npm &> /dev/null; then
        echo "Checking for outdated mobile packages..."
        npm outdated || true

        echo ""
        echo "Updating mobile packages..."
        npm update

        print_success "Mobile dependencies updated"
    else
        print_error "npm not found, skipping mobile dependencies"
    fi
else
    print_warning "Mobile directory not found, skipping"
fi

echo ""
echo "4. Running tests..."
echo "==================="

# Run Go tests
cd /home/user/marimo
echo "Running Go tests..."
if go test ./... -short; then
    print_success "Go tests passed"
else
    print_error "Go tests failed - please review changes"
    exit 1
fi

# Run frontend tests
if [ -d "/home/user/marimo/frontend" ]; then
    cd /home/user/marimo/frontend
    echo ""
    echo "Running frontend tests..."
    if npm test -- --watchAll=false --passWithNoTests; then
        print_success "Frontend tests passed"
    else
        print_error "Frontend tests failed - please review changes"
        exit 1
    fi
fi

echo ""
echo "5. Summary"
echo "=========="
echo ""
print_success "All dependencies updated successfully!"
echo ""
echo "Next steps:"
echo "1. Review the changes: git diff go.mod go.sum frontend/package.json"
echo "2. Test the application thoroughly"
echo "3. Commit the changes: git add . && git commit -m 'chore: update dependencies'"
echo "4. Push to remote: git push"
echo ""
echo "====================================="
