#!/usr/bin/env bash
set -e
cd "$(dirname "$0")"

command -v go >/dev/null || { echo "Go is required to build: sudo apt install golang-go"; exit 1; }

BIN="$HOME/.local/bin"
APPS="$HOME/.local/share/applications"
MIME="$HOME/.local/share/mime"
mkdir -p "$BIN" "$APPS" "$MIME/packages"

go build -o "$BIN/castplay" .
install -m 755 castplay-launch "$BIN/castplay-launch"

cat > "$MIME/packages/asciicast.xml" << XMLEOF
<?xml version="1.0" encoding="UTF-8"?>
<mime-info xmlns="http://www.freedesktop.org/standards/shared-mime-info">
  <mime-type type="application/x-asciicast">
    <comment>asciinema recording</comment>
    <glob pattern="*.cast"/>
  </mime-type>
</mime-info>
XMLEOF
update-mime-database "$MIME" || true

cat > "$APPS/castplay.desktop" << DESKTOPEOF
[Desktop Entry]
Type=Application
Name=castplay
Exec=$BIN/castplay-launch %f
Terminal=true
MimeType=application/x-asciicast;
NoDisplay=true
DESKTOPEOF
update-desktop-database "$APPS" || true

gio mime application/x-asciicast castplay.desktop || echo "Could not set default app; use Open With on a .cast file."

echo "Installed to $BIN. Make sure $BIN is on your PATH."
