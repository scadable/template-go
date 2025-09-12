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
# Install yq (https://github.com/mikefarah/yq) if not installed
# -----------------------------------------------------------------------------
if ! command -v yq &> /dev/null; then
  echo "🔧 Installing yq..."
  OS="$(uname -s)"
  ARCH="$(uname -m)"

  case "$OS" in
    Darwin)
      brew install yq || { echo "🍺 Homebrew not found. Please install yq manually."; exit 1; }
      ;;
    Linux)
      YQ_BIN="yq_linux_amd64"
      if [ "$ARCH" = "aarch64" ]; then
        YQ_BIN="yq_linux_arm64"
      fi
      sudo wget -q "https://github.com/mikefarah/yq/releases/latest/download/${YQ_BIN}" -O /usr/local/bin/yq
      sudo chmod +x /usr/local/bin/yq
      ;;
    *)
      echo "❌ Unsupported OS: $OS. Please install yq manually from https://github.com/mikefarah/yq"
      exit 1
      ;;
  esac
else
  echo "✅ yq already installed"
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
