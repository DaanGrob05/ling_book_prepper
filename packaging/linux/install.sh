#!/bin/sh
# Installs LingQ Prep for the current user, without needing sudo: the
# program goes to ~/.local/bin and the app menu entry and icon to
# ~/.local/share. Run it from the extracted release folder: ./install.sh
set -eu
cd "$(dirname "$0")"

id=io.github.daangrob05.lingqprep
bin="$HOME/.local/bin"
share="${XDG_DATA_HOME:-$HOME/.local/share}"

install -Dm755 usr/local/bin/lingqprep "$bin/lingqprep"
install -Dm644 "usr/local/share/pixmaps/$id.png" "$share/icons/hicolor/512x512/apps/$id.png"
# The menu entry must point to where the program now is.
mkdir -p "$share/applications"
sed "s|^Exec=lingqprep|Exec=$bin/lingqprep|" "usr/local/share/applications/$id.desktop" \
	> "$share/applications/$id.desktop"
update-desktop-database "$share/applications" 2>/dev/null || true

echo "LingQ Prep is installed. You can find it in your app menu."
case ":$PATH:" in
*":$bin:"*) echo "The lingqprep command works in a terminal too." ;;
*) echo "To use the lingqprep command in a terminal, add $bin to your PATH." ;;
esac
