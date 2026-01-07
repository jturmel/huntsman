#!/usr/bin/env bash

# Exit on error
set -e

# Define paths
INSTALL_DIR="$HOME/.local/bin"
BINARY_NAME="huntsman"
REPO="jturmel/huntsman"
VERSION="${VERSION:-latest}"

# Detect OS and Architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Error: Unsupported architecture $ARCH"
        exit 1
        ;;
esac

case $OS in
    linux)
        PLATFORM="linux"
        ;;
    darwin)
        PLATFORM="darwin"
        ;;
    *)
        echo "Error: Unsupported OS $OS"
        exit 1
        ;;
esac

ASSET_NAME="${BINARY_NAME}-${PLATFORM}-${ARCH}"

echo "Installing $BINARY_NAME ($VERSION) for ${PLATFORM}-${ARCH}..."

# Function to download from GitHub release
download_from_release() {
    local file=$1
    if ! command -v curl >/dev/null 2>&1; then
        echo "Error: curl is not installed. Cannot download missing files."
        exit 1
    fi
    echo "Downloading $file from release $VERSION..."
    curl -sL --fail "https://github.com/$REPO/releases/download/$VERSION/$file" -o "$BINARY_NAME" || {
        echo "Error: Failed to download $file. It might not be available in the $VERSION release."
        exit 1
    }
}

# 1. Setup directory in ~/.local/bin
echo "Creating directory $INSTALL_DIR..."
mkdir -p "$INSTALL_DIR"

# 2. Get binary
if [ -f "$ASSET_NAME" ]; then
    echo "Using local binary $ASSET_NAME..."
    cp "$ASSET_NAME" "$BINARY_NAME"
elif [ -f "$BINARY_NAME" ]; then
    echo "Using local binary $BINARY_NAME..."
else
    download_from_release "$ASSET_NAME"
fi

# Check if the binary exists now
if [ ! -f "$BINARY_NAME" ]; then
    echo "Error: $BINARY_NAME not found and failed to download."
    exit 1
fi

# 3. Copy the binary
echo "Copying binary to $INSTALL_DIR..."
cp "$BINARY_NAME" "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/$BINARY_NAME"
rm "$BINARY_NAME"

echo "Installation complete!"
echo "Note: Make sure $INSTALL_DIR is in your PATH."
