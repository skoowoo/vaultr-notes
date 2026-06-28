#!/usr/bin/env bash
# Generate icon.icns from icon.png (macOS only).
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
PNG="$DIR/icon.png"
ICNS="$DIR/icon.icns"
ICONSET="$DIR/icon.iconset"

if [[ "$(uname)" != "Darwin" ]]; then
  echo "error: iconutil requires macOS" >&2
  exit 1
fi

if [[ ! -f "$PNG" ]]; then
  echo "error: missing $PNG" >&2
  exit 1
fi

rm -rf "$ICONSET"
mkdir -p "$ICONSET"

sips -z 16  16  "$PNG" --out "$ICONSET/icon_16x16.png"      >/dev/null
sips -z 32  32  "$PNG" --out "$ICONSET/icon_16x16@2x.png"  >/dev/null
sips -z 32  32  "$PNG" --out "$ICONSET/icon_32x32.png"     >/dev/null
sips -z 64  64  "$PNG" --out "$ICONSET/icon_32x32@2x.png"  >/dev/null
sips -z 128 128 "$PNG" --out "$ICONSET/icon_128x128.png"    >/dev/null
sips -z 256 256 "$PNG" --out "$ICONSET/icon_128x128@2x.png" >/dev/null
sips -z 256 256 "$PNG" --out "$ICONSET/icon_256x256.png"    >/dev/null
sips -z 512 512 "$PNG" --out "$ICONSET/icon_256x256@2x.png" >/dev/null
sips -z 512 512 "$PNG" --out "$ICONSET/icon_512x512.png"    >/dev/null
cp "$PNG" "$ICONSET/icon_512x512@2x.png"

iconutil -c icns "$ICONSET" -o "$ICNS"
rm -rf "$ICONSET"

echo "generated: $ICNS ($(ls -lh "$ICNS" | awk '{print $5}'))"
