#!/bin/bash
set -e

REPO="jagtesh/guise"
BINARY="guise"
INSTALL_DIR="/usr/local/bin"

echo "Installing Guise..."

# Determine OS and Arch
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

if [ "$ARCH" == "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" == "aarch64" ] || [ "$ARCH" == "arm64" ]; then
    ARCH="arm64"
else
    echo "Unsupported architecture: $ARCH"
    exit 1
fi

# In a real scenario, we would download from GitHub Releases.
# Since this is a local project simulation, we'll assume 'go install' or build.
# For this script to be valid in the repo, I'll write the logic as if downloading a release.

echo "Downloading latest release..."
# URL="https://github.com/$REPO/releases/latest/download/guise-$OS-$ARCH"
# curl -L $URL -o $BINARY
# chmod +x $BINARY
# sudo mv $BINARY $INSTALL_DIR/

# For now, build locally if Go is present, otherwise fail gracefully instructions
if command -v go &> /dev/null; then
    echo "Go detected. Building from source..."
    go build -o guise main.go
    echo "Moving to $INSTALL_DIR (requires sudo)..."
    sudo mv guise "$INSTALL_DIR/"
    echo "Success! Run 'guise' to start."
else
    echo "Error: Pre-built binaries not yet hosted. Please install Go to build from source."
    exit 1
fi
