#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SVG="$SCRIPT_DIR/icon.svg"
PNG="$SCRIPT_DIR/icon.png"
SIZE=1024

rsvg-convert -w $SIZE -h $SIZE -o "$PNG" "$SVG"

echo "Exported: $PNG (${SIZE}x${SIZE})"
