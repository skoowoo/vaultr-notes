#!/bin/sh
set -e

REPO="skoowoo/vaultr-notes"
REPO_BRANCH="${REPO_BRANCH:-main}"
SKILLS_DIR="$HOME/.vaultr/skills"
EXTERNAL_SKILLS_FILE=""

# Parse arguments
while [ $# -gt 0 ]; do
    case "$1" in
        --external-skills-file)
            shift
            EXTERNAL_SKILLS_FILE="$1"
            ;;
        --external-skills-file=*)
            EXTERNAL_SKILLS_FILE="${1#*=}"
            ;;
        --skills-dir)
            shift
            SKILLS_DIR="$1"
            ;;
        --skills-dir=*)
            SKILLS_DIR="${1#*=}"
            ;;
        --help|-h)
            echo "Usage: $0 [--external-skills-file FILE] [--skills-dir DIR]"
            echo ""
            echo "Install external skills listed in external-skills.txt to ~/.vaultr/skills/."
            echo ""
            echo "Options:"
            echo "  --external-skills-file FILE  Use a local external-skills.txt (default: fetch latest from repo)"
            echo "  --skills-dir DIR             Destination skills directory (default: ~/.vaultr/skills)"
            exit 0
            ;;
        *)
            echo "Error: Unknown argument: $1"
            echo "Usage: $0 [--external-skills-file FILE] [--skills-dir DIR]"
            exit 1
            ;;
    esac
    shift
done

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

mkdir -p "$SKILLS_DIR"

if [ -z "$EXTERNAL_SKILLS_FILE" ]; then
    EXTERNAL_SKILLS_URL="https://raw.githubusercontent.com/$REPO/$REPO_BRANCH/external-skills.txt"
    EXTERNAL_SKILLS_FILE="$TMP_DIR/external-skills.txt"
    echo "==> Fetching latest external-skills.txt from $EXTERNAL_SKILLS_URL..."
    if ! curl -sSL -f -o "$EXTERNAL_SKILLS_FILE" "$EXTERNAL_SKILLS_URL"; then
        echo "${RED}Error: Failed to fetch external-skills.txt from $EXTERNAL_SKILLS_URL${NC}"
        exit 1
    fi
elif [ ! -f "$EXTERNAL_SKILLS_FILE" ]; then
    echo "${RED}Error: external-skills.txt not found at $EXTERNAL_SKILLS_FILE${NC}"
    exit 1
fi

echo "==> Installing external skills from $EXTERNAL_SKILLS_FILE..."
while IFS= read -r entry || [ -n "$entry" ]; do
    # Skip blank lines and comment lines
    case "$entry" in
        ""|\#*) continue ;;
    esac

    # Parse three-segment format: "GitHubUser/RepoName:path/in/repo:skill-dir-name"
    # An empty path (e.g. user/repo::skill-name) or "." means the repo root.
    GITHUB_PATH="${entry%%:*}"
    _rest="${entry#*:}"
    REPO_SUBPATH="${_rest%%:*}"
    SKILL_NAME="${_rest##*:}"

    if [ -z "$GITHUB_PATH" ] || [ -z "$SKILL_NAME" ]; then
        echo "${YELLOW}Warning: Skipping invalid external skill entry: $entry${NC}"
        echo "${YELLOW}         Expected format: GitHubUser/RepoName:path/in/repo:skill-dir-name${NC}"
        continue
    fi

    case "$REPO_SUBPATH" in
        "") SKILL_SRC_LABEL="repo root" ;;
        ".") REPO_SUBPATH="" ; SKILL_SRC_LABEL="repo root" ;;
        *) SKILL_SRC_LABEL="$REPO_SUBPATH" ;;
    esac

    SKILL_URL="https://github.com/${GITHUB_PATH}.git"
    SKILL_DEST="$SKILLS_DIR/$SKILL_NAME"

    echo "==> Installing external skill '$SKILL_NAME' from $SKILL_URL ($SKILL_SRC_LABEL)..."

    SKILL_TMP="$TMP_DIR/ext_skill_$SKILL_NAME"
    if git clone --depth=1 --quiet "$SKILL_URL" "$SKILL_TMP" 2>/dev/null; then
        if [ -z "$REPO_SUBPATH" ]; then
            SKILL_SRC="$SKILL_TMP"
        else
            SKILL_SRC="$SKILL_TMP/$REPO_SUBPATH"
        fi

        if [ -d "$SKILL_SRC" ]; then
            if [ -L "$SKILL_DEST" ]; then
                # $SKILL_DEST is a symlink — update the symlink target in-place to preserve the link
                REAL_DEST=$(realpath "$SKILL_DEST" 2>/dev/null || readlink "$SKILL_DEST")
                rm -rf "${REAL_DEST:?}/"
                mkdir -p "$REAL_DEST"
                if [ -z "$REPO_SUBPATH" ]; then
                    (cd "$SKILL_SRC" && tar cf - --exclude=.git .) | (cd "$REAL_DEST" && tar xf -)
                else
                    cp -r "$SKILL_SRC/." "$REAL_DEST/"
                fi
                echo "==> Skill '$SKILL_NAME' updated via symlink -> $REAL_DEST"
            else
                rm -rf "$SKILL_DEST"
                if [ -z "$REPO_SUBPATH" ]; then
                    mkdir -p "$SKILL_DEST"
                    (cd "$SKILL_SRC" && tar cf - --exclude=.git .) | (cd "$SKILL_DEST" && tar xf -)
                else
                    cp -r "$SKILL_SRC" "$SKILL_DEST"
                fi
                echo "==> Skill '$SKILL_NAME' installed to $SKILL_DEST"
            fi
        else
            echo "${YELLOW}Warning: Path '$SKILL_SRC_LABEL' not found in $SKILL_URL, skipping.${NC}"
        fi
    else
        echo "${YELLOW}Warning: Failed to clone $SKILL_URL, skipping skill '$SKILL_NAME'.${NC}"
    fi
done < "$EXTERNAL_SKILLS_FILE"

cp "$EXTERNAL_SKILLS_FILE" "$SKILLS_DIR/external-skills.txt"
echo "==> Saved external-skills.txt to $SKILLS_DIR/"

echo "${GREEN}==> External skills installation complete!${NC}"
