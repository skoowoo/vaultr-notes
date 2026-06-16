#!/bin/sh
set -e

REPO="skoowoo/vaultr-notes"
BINARY_NAME="vaultr"

# Default install directory
INSTALL_DIR="/usr/local/bin"

# Parse arguments
SKILLS_ONLY=0
for arg in "$@"; do
    case "$arg" in
        --skills-only) SKILLS_ONLY=1 ;;
        *) echo "${RED}Error: Unknown argument: $arg${NC}"; echo "Usage: $0 [--skills-only]"; exit 1 ;;
    esac
done

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Create a temporary directory (used by both full install and skills-only)
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

# Set up ~/.vaultr directory
VAULTR_DIR="$HOME/.vaultr"
mkdir -p "$VAULTR_DIR"

if [ "$SKILLS_ONLY" = "0" ]; then
    echo "==> Installing $BINARY_NAME from $REPO..."

    # Detect OS
    OS=$(uname -s)
    case "$OS" in
        Linux)  OS="linux" ;;
        Darwin) OS="darwin" ;;
        *)      echo "${RED}Error: Unsupported operating system: $OS${NC}"; exit 1 ;;
    esac

    # Detect Architecture
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64|amd64) ARCH="amd64" ;;
        arm64|aarch64) ARCH="arm64" ;;
        *)           echo "${RED}Error: Unsupported architecture: $ARCH${NC}"; exit 1 ;;
    esac

    echo "==> Detected OS: $OS, Architecture: $ARCH"

    # Fetch latest version if not specified
    if [ -z "$VERSION" ]; then
        echo "==> Fetching latest release version..."
        VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
        if [ -z "$VERSION" ]; then
            echo "${RED}Error: Could not determine the latest version from GitHub.${NC}"
            exit 1
        fi
    fi

    echo "==> Version to install: $VERSION"

    # Format the filename: vaultr_linux_amd64.tar.gz
    FILE_NAME="${BINARY_NAME}_${OS}_${ARCH}.tar.gz"
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/$FILE_NAME"

    echo "==> Downloading $FILE_NAME..."
    curl -sL -o "$TMP_DIR/$FILE_NAME" "$DOWNLOAD_URL"

    echo "==> Extracting archive..."
    cd "$TMP_DIR"
    tar xzf "$FILE_NAME"

    if [ ! -f "$BINARY_NAME" ]; then
        echo "${RED}Error: Binary '$BINARY_NAME' not found in the downloaded archive.${NC}"
        exit 1
    fi

    # Determine the installation directory (fallback to ~/.local/bin if /usr/local/bin is not writable)
    if [ ! -w "$INSTALL_DIR" ]; then
        INSTALL_DIR="$HOME/.local/bin"
        mkdir -p "$INSTALL_DIR"

        # Check if ~/.local/bin is in PATH
        if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
            echo "${YELLOW}Warning: $INSTALL_DIR is not in your PATH.${NC}"
            echo "Please add the following line to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
            echo "export PATH=\"\$PATH:$INSTALL_DIR\""
        fi
    fi

    echo "==> Installing to $INSTALL_DIR..."
    mv "$BINARY_NAME" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"

    if [ -f "$TMP_DIR/config.example.toml" ]; then
        if [ ! -f "$VAULTR_DIR/config.toml" ]; then
            echo "==> Installing default config to $VAULTR_DIR/config.toml..."
            cp "$TMP_DIR/config.example.toml" "$VAULTR_DIR/config.toml"
        else
            echo "==> Config already exists at $VAULTR_DIR/config.toml, skipping."
        fi
    fi

    if [ -d "$TMP_DIR/skills" ]; then
        echo "==> Installing built-in skills to $VAULTR_DIR/skills/..."
        cp -r "$TMP_DIR/skills" "$VAULTR_DIR/"
    fi
else
    echo "==> Skills-only mode: skipping binary installation."
fi

# Install external skills from GitHub
SKILLS_DIR="$VAULTR_DIR/skills"
mkdir -p "$SKILLS_DIR"

# In skills-only mode, read external-skills.txt from the script's own directory (for local testing).
# In normal mode, read it from the extracted archive in TMP_DIR.
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
if [ "$SKILLS_ONLY" = "1" ]; then
    EXTERNAL_SKILLS_FILE="$SCRIPT_DIR/external-skills.txt"
else
    EXTERNAL_SKILLS_FILE="$TMP_DIR/external-skills.txt"
fi
if [ ! -f "$EXTERNAL_SKILLS_FILE" ]; then
    echo "==> No external-skills.txt found, skipping external skills."
else
    echo "==> Installing external skills from external-skills.txt..."
    while IFS= read -r entry || [ -n "$entry" ]; do
        # Skip blank lines and comment lines
        case "$entry" in
            ""|\#*) continue ;;
        esac

        # Parse three-segment format: "GitHubUser/RepoName:path/in/repo:skill-dir-name"
        GITHUB_PATH="${entry%%:*}"
        _rest="${entry#*:}"
        REPO_SUBPATH="${_rest%%:*}"
        SKILL_NAME="${_rest##*:}"

        if [ -z "$GITHUB_PATH" ] || [ -z "$REPO_SUBPATH" ] || [ -z "$SKILL_NAME" ]; then
            echo "${YELLOW}Warning: Skipping invalid external skill entry: $entry${NC}"
            echo "${YELLOW}         Expected format: GitHubUser/RepoName:path/in/repo:skill-dir-name${NC}"
            continue
        fi

        SKILL_URL="https://github.com/${GITHUB_PATH}.git"
        SKILL_DEST="$SKILLS_DIR/$SKILL_NAME"

        echo "==> Installing external skill '$SKILL_NAME' from $SKILL_URL ($REPO_SUBPATH)..."

        SKILL_TMP="$TMP_DIR/ext_skill_$SKILL_NAME"
        if git clone --depth=1 --quiet "$SKILL_URL" "$SKILL_TMP" 2>/dev/null; then
            if [ -d "$SKILL_TMP/$REPO_SUBPATH" ]; then
                if [ -L "$SKILL_DEST" ]; then
                    # $SKILL_DEST is a symlink — update the symlink target in-place to preserve the link
                    REAL_DEST=$(realpath "$SKILL_DEST" 2>/dev/null || readlink "$SKILL_DEST")
                    rm -rf "${REAL_DEST:?}/"
                    cp -r "$SKILL_TMP/$REPO_SUBPATH/." "$REAL_DEST/"
                    echo "==> Skill '$SKILL_NAME' updated via symlink -> $REAL_DEST"
                else
                    rm -rf "$SKILL_DEST"
                    cp -r "$SKILL_TMP/$REPO_SUBPATH" "$SKILL_DEST"
                    echo "==> Skill '$SKILL_NAME' installed to $SKILL_DEST"
                fi
            else
                echo "${YELLOW}Warning: Path '$REPO_SUBPATH' not found in $SKILL_URL, skipping.${NC}"
            fi
        else
            echo "${YELLOW}Warning: Failed to clone $SKILL_URL, skipping skill '$SKILL_NAME'.${NC}"
        fi
    done < "$EXTERNAL_SKILLS_FILE"
    cp "$EXTERNAL_SKILLS_FILE" "$SKILLS_DIR/external-skills.txt"
    echo "==> Saved external-skills.txt to $SKILLS_DIR/"
fi

echo "${GREEN}==> Installation complete!${NC}"
echo "You can now run: $BINARY_NAME"
