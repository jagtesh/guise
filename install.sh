#!/bin/bash
set -e

REPO="jagtesh/guise"
BINARY="guise"
INSTALL_DIR="/usr/local/bin"

# Detect OS and Arch
OS="$(uname -s)"
ARCH="$(uname -m)"

# Normalize Arch
if [ "$ARCH" == "x86_64" ]; then
    ARCH="x86_64"
elif [ "$ARCH" == "aarch64" ] || [ "$ARCH" == "arm64" ]; then
    ARCH="arm64"
else
    echo "Unsupported architecture: $ARCH"
    exit 1
fi

# Normalize OS
if [ "$OS" == "Linux" ]; then
    OS="Linux"
elif [ "$OS" == "Darwin" ]; then
    OS="Darwin"
else
    echo "Unsupported OS: $OS"
    exit 1
fi

FILE="${BINARY}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/$REPO/releases/latest/download/$FILE"

echo "Downloading Guise ($OS $ARCH)..."
echo "URL: $DOWNLOAD_URL"

# Create temp directory
TMP_DIR=$(mktemp -d)
curl -sL "$DOWNLOAD_URL" -o "$TMP_DIR/$FILE"

echo "Installing..."
tar -xzf "$TMP_DIR/$FILE" -C "$TMP_DIR"
chmod +x "$TMP_DIR/$BINARY"

# Move to install dir (requires sudo usually)
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/"
else
    echo "Sudo required to move binary to $INSTALL_DIR"
    sudo mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/"
fi

# Cleanup
rm -rf "$TMP_DIR"

echo "Success! Run '$BINARY' to start."