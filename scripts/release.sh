#!/bin/bash
set -e

VERSION="$1"

if [ -z "$VERSION" ]; then
    echo "Usage: ./release.sh <version>  (e.g. ./release.sh v1.0.0)"
    exit 1
fi

echo "→ Releasing $VERSION"

# Create and push git tag
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    echo "  Tag $VERSION already exists, skipping tag creation"
else
    git tag "$VERSION"
    git push origin "$VERSION"
    echo "  Tag $VERSION pushed"
fi

# Collect dist artifacts (dmg, tar.gz, zip, checksums)
ASSETS=()
for pattern in "./dist/*.dmg" "./dist/*.tar.gz" "./dist/*.zip" "./dist/checksums.txt"; do
    for f in $pattern; do
        [ -f "$f" ] && ASSETS+=("$f")
    done
done

if [ ${#ASSETS[@]} -eq 0 ]; then
    echo "Error: no release assets found in ./dist/"
    exit 1
fi

echo "  Assets to upload:"
for f in "${ASSETS[@]}"; do
    echo "    $f"
done

# Create GitHub release
gh release create "$VERSION" \
    "${ASSETS[@]}" \
    --generate-notes \
    --title "$VERSION" \
    --latest

echo "✓ Release $VERSION published"
