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

sha256_check() {
    if command -v sha256sum > /dev/null 2>&1; then
        sha256sum "$1" | awk '{print $1}'
    elif command -v shasum > /dev/null 2>&1; then
        shasum -a 256 "$1" | awk '{print $1}'
    else
        echo ""
    fi
}

OS=$(get_os)
ARCH=$(get_arch)

echo "Detecting system: ${OS}/${ARCH}"

if [ -n "$1" ]; then
    VERSION="$1"
else
    VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed 's/.*"tag_name": *"//;s/".*//')
    if [ -z "$VERSION" ]; then
        echo "No releases found. You can specify a version: ./install.sh v0.1.0"
        exit 1
    fi
fi

ASSET="${BINARY_NAME}-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"
CHECKSUMS_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"

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

# Verify SHA256 checksum
echo "Verifying checksum..."
CHECKSUMS_TMP=$(mktemp)
DOWNLOAD_OK=true
if command -v curl > /dev/null 2>&1; then
    curl -fSL -o "$CHECKSUMS_TMP" "$CHECKSUMS_URL" 2>/dev/null || DOWNLOAD_OK=false
elif command -v wget > /dev/null 2>&1; then
    wget -q -O "$CHECKSUMS_TMP" "$CHECKSUMS_URL" 2>/dev/null || DOWNLOAD_OK=false
fi

if [ "$DOWNLOAD_OK" = true ] && [ -s "$CHECKSUMS_TMP" ]; then
    EXPECTED=$(grep "${ASSET}" "$CHECKSUMS_TMP" | awk '{print $1}')
    ACTUAL=$(sha256_check "$TMP")
    if [ -n "$EXPECTED" ] && [ -n "$ACTUAL" ]; then
        if [ "$EXPECTED" != "$ACTUAL" ]; then
            echo "Checksum verification FAILED!"
            echo "  Expected: ${EXPECTED}"
            echo "  Got:      ${ACTUAL}"
            rm -f "$TMP" "$CHECKSUMS_TMP"
            exit 1
        fi
        echo "Checksum verified OK."
    else
        echo "Warning: could not verify checksum (sha256sum/shasum not available)."
    fi
else
    echo "Warning: checksums.txt not available for this release, skipping verification."
fi
rm -f "$CHECKSUMS_TMP"

chmod +x "$TMP"

if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP" "${INSTALL_DIR}/${BINARY_NAME}"
else
    echo "Installing to ${INSTALL_DIR} (requires sudo)..."
    sudo mv "$TMP" "${INSTALL_DIR}/${BINARY_NAME}"
fi

echo "CColon ${VERSION} installed to ${INSTALL_DIR}/${BINARY_NAME}"
echo "Run 'ccolon' to start the interactive shell."
