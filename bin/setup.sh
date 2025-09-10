#!/bin/bash

# -----------------------------------------------------------------------------
# GoLang Development Environment Setup Script
# Run this from the root of the project: ./bin/setup.sh
# -----------------------------------------------------------------------------

# Start
echo "🚀 Setting up GoLang development environment..."
echo "⏳ This may take a moment..."

# -----------------------------------------------------------------------------
# Ensure we're running from the project root
# -----------------------------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT" || exit 1

# -----------------------------------------------------------------------------
# Install golangci-lint if not already installed
# -----------------------------------------------------------------------------
if ! command -v golangci-lint &> /dev/null; then
  echo "🔧 Installing golangci-lint..."
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
else
  echo "✅ golangci-lint already installed"
fi

# -----------------------------------------------------------------------------
# Run go mod tidy to sync dependencies
# -----------------------------------------------------------------------------
echo "📦 Running go mod tidy..."
go mod tidy

# -----------------------------------------------------------------------------
# Done
# -----------------------------------------------------------------------------
echo "✅ Environment setup complete."
