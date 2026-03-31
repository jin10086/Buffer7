#!/bin/bash
set -e

# Repository configuration
REPO="jin10086/buffer7"
BINARY_NAME="buffer7"

# Check dependencies
for cmd in curl tar uname; do
    if ! command -v $cmd >/dev/null 2>&1; then
        echo "Error: $cmd is required but not installed."
        exit 1
    fi
done

# Version discovery
echo "Discovering latest version..."
LATEST_TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_TAG" ]; then
    echo "Error: Could not find the latest version. Please check your internet connection or repository URL."
    exit 1
fi
echo "Latest version found: $LATEST_TAG"

# Platform detection
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Error: Unsupported architecture $ARCH"; exit 1 ;;
esac

case $OS in
    linux|darwin) ;;
    *) echo "Error: Unsupported OS $OS"; exit 1 ;;
esac

FILE_EXT="tar.gz"
# GoReleaser naming: {{ .ProjectName }}_{{ .Os }}_{{ .Arch }}
FILENAME="${BINARY_NAME}_${OS}_${ARCH}.${FILE_EXT}"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_TAG/$FILENAME"

# Download and extract
TEMP_DIR=$(mktemp -d)
trap 'rm -rf "$TEMP_DIR"' EXIT

echo "Downloading $FILENAME..."
if ! curl -L -o "$TEMP_DIR/$FILENAME" "$DOWNLOAD_URL"; then
    echo "Error: Failed to download $DOWNLOAD_URL"
    exit 1
fi

echo "Extracting binary..."
tar -xzf "$TEMP_DIR/$FILENAME" -C "$TEMP_DIR"

# Install binary
INSTALL_DIR="/usr/local/bin"
echo "Installing $BINARY_NAME to $INSTALL_DIR..."

if [ -w "$INSTALL_DIR" ]; then
    mv "$TEMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
else
    echo "Permission denied for $INSTALL_DIR. Using sudo..."
    sudo mv "$TEMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
fi

chmod +x "$INSTALL_DIR/$BINARY_NAME"

# Verification
echo "Verifying installation..."
if command -v "$BINARY_NAME" >/dev/null 2>&1; then
    VERSION=$($BINARY_NAME --version 2>&1 || true)
    echo "Success! $BINARY_NAME is installed: $VERSION"
else
    echo "Error: $BINARY_NAME installation failed or is not in PATH."
    exit 1
fi
