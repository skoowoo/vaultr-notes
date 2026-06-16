#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SVG="$SCRIPT_DIR/icon.svg"
OUT_DIR="$SCRIPT_DIR/../../extensions/clip/icons"

for SIZE in 16 48 128; do
  rsvg-convert -w $SIZE -h $SIZE -o "$OUT_DIR/icon${SIZE}.png" "$SVG"
  echo "Exported: $OUT_DIR/icon${SIZE}.png (${SIZE}x${SIZE})"
done
