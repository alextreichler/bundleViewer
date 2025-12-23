#!/bin/bash
set -e

# --- Configuration ---
REPO="alextreichler/bundleViewer"
BINARY_NAME="bundleViewer"
INSTALL_DIR="/usr/local/bin"

# --- Colors ---
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}>>> Starting installation for ${BINARY_NAME}...${NC}"

# --- 1. Detect Architecture & OS ---
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map architecture to the naming convention used in your Taskfile/Releases
case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) 
        echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"
        exit 1 
        ;;
esac

# Construct the expected asset name (Must match what you upload to GitHub Releases)
# Based on your Taskfile: bundleViewer-linux-amd64 or bundleViewer-darwin-arm64
ASSET_NAME="${BINARY_NAME}-${OS}-${ARCH}"

echo -e "Detected Platform: ${GREEN}${OS}/${ARCH}${NC}"

# --- 2. Determine Download URL ---
# We use the 'latest' release endpoint.
DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${ASSET_NAME}"

# --- 3. Download ---
echo -e "Downloading from: ${BLUE}${DOWNLOAD_URL}${NC}"

# Create a temporary file
TMP_FILE=$(mktemp)

if curl --fail -L --progress-bar "$DOWNLOAD_URL" -o "$TMP_FILE"; then
    echo -e "${GREEN}Download successful.${NC}"
else
    echo -e "${RED}Error: Download failed.${NC}"
    echo "Check if a release exists for '${ASSET_NAME}' at https://github.com/${REPO}/releases"
    rm -f "$TMP_FILE"
    exit 1
fi

# --- 4. Install ---
echo "Installing to ${INSTALL_DIR}..."

# Check if we have write access to INSTALL_DIR, otherwise use sudo
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
else
    echo "Sudo permissions required to write to ${INSTALL_DIR}"
    sudo mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
    sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
fi

# --- 5. Verify ---
if command -v "$BINARY_NAME" >/dev/null; then
    INSTALLED_PATH=$(which "$BINARY_NAME")
    echo -e "${GREEN}Success! ${BINARY_NAME} installed to ${INSTALLED_PATH}${NC}"
    echo -e "Run '${BLUE}${BINARY_NAME} --help${NC}' to get started."
else
    echo -e "${RED}Error: Installation appeared to succeed, but binary not found in PATH.${NC}"
    echo "Please ensure ${INSTALL_DIR} is in your PATH."
fi