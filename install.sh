#!/bin/sh
set -e

REPO="TRC-Loop/CColon"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="ccolon"

get_os() {
    case "$(uname -s)" in
        Linux*)  echo "linux" ;;
        Darwin*) echo "darwin" ;;
        *)       echo "unsupported"; exit 1 ;;
    esac
}

get_arch() {
    case "$(uname -m)" in
        x86_64|amd64)  echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *)             echo "unsupported"; exit 1 ;;
    esac
}

OS=$(get_os)
ARCH=$(get_arch)

echo "Detecting system: ${OS}/${ARCH}"

if [ -n "$1" ]; then
    VERSION="$1"
else
    VERSION=$(curl -sI "https://github.com/${REPO}/releases/latest" | grep -i "location:" | sed 's/.*tag\///' | tr -d '\r\n')
    if [ -z "$VERSION" ]; then
        echo "Failed to determine latest version. You can specify a version: ./install.sh v0.1.0"
        exit 1
    fi
fi

URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}-${OS}-${ARCH}"

echo "Downloading CColon ${VERSION} from ${URL}..."

TMP=$(mktemp)
if command -v curl > /dev/null 2>&1; then
    curl -fSL -o "$TMP" "$URL"
elif command -v wget > /dev/null 2>&1; then
    wget -q -O "$TMP" "$URL"
else
    echo "curl or wget is required to download CColon."
    exit 1
fi

chmod +x "$TMP"

if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP" "${INSTALL_DIR}/${BINARY_NAME}"
else
    echo "Installing to ${INSTALL_DIR} (requires sudo)..."
    sudo mv "$TMP" "${INSTALL_DIR}/${BINARY_NAME}"
fi

echo "CColon ${VERSION} installed to ${INSTALL_DIR}/${BINARY_NAME}"
echo "Run 'ccolon' to start the interactive shell."
