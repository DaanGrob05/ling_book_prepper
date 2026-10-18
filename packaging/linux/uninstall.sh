#!/bin/sh
# Removes what install.sh installed.
set -eu

id=io.github.daangrob05.lingqprep
share="${XDG_DATA_HOME:-$HOME/.local/share}"

rm -f "$HOME/.local/bin/lingqprep" \
	"$share/icons/hicolor/512x512/apps/$id.png" \
	"$share/applications/$id.desktop"
update-desktop-database "$share/applications" 2>/dev/null || true

echo "LingQ Prep is uninstalled."
