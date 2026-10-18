#!/bin/sh
# Builds the Linux release archive, e.g.
#
#	packaging/linux/package.sh 1.2.0-beta.1 dist
#
# writes dist/LingQ-Prep-1.2.0-beta.1-Linux-x86_64.tar.xz. The version is
# the git tag without its "v"; the commit being built must have that tag,
# which is checked. Needs the fyne tool (go install fyne.io/tools/cmd/fyne).
#
# The binary must run on older systems too, so it is checked to need at
# most glibc $MAX_GLIBC (default 2.35, Ubuntu 22.04). Build releases on
# Ubuntu 22.04 to pass the check.
set -eu

version=$1
mkdir -p "$2"
out=$(cd "$2" && pwd)
max_glibc=${MAX_GLIBC:-2.35}
root=$(cd "$(dirname "$0")/../.." && pwd)
name="LingQ-Prep-$version"

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# fyne package rewrites FyneApp.toml (it raises Build), so restore it after.
cd "$root/cmd/lingqprep"
cp FyneApp.toml "$work/FyneApp.toml"
fyne package --os linux --release \
	--app-version "${version%%-*}" --app-build "${BUILD_NUMBER:-1}"
cp "$work/FyneApp.toml" FyneApp.toml
tar -xJf "LingQ Prep.tar.xz" -C "$work"
rm "LingQ Prep.tar.xz"

# fyne package lays the files out for "make install" (usr/local/...). Use a
# flat folder instead, with the program at the top where people look for it.
fyne="$work/lingqprep/usr/local"
dir="$work/$name"
mkdir -p "$dir/share"
cp "$fyne/bin/lingqprep" "$dir/"
cp "$fyne/share/applications/"*.desktop "$fyne/share/pixmaps/"*.png "$dir/share/"
cp "$root/packaging/linux/install.sh" "$root/packaging/linux/uninstall.sh" \
	"$root/packaging/linux/README.txt" "$dir/"
bin="$dir/lingqprep"

needed=$(objdump -T "$bin" | grep -o 'GLIBC_[0-9.]*' | sed 's/GLIBC_//' | sort -uV | tail -n 1)
if [ "$(printf '%s\n%s\n' "$needed" "$max_glibc" | sort -V | tail -n 1)" != "$max_glibc" ]; then
	echo "error: the binary needs glibc $needed, newer than $max_glibc; build on an older system" >&2
	exit 1
fi
echo "binary needs glibc $needed (limit $max_glibc)"

got=$("$bin" -version)
if [ "$got" != "lingqprep v$version" ]; then
	echo "error: the binary reports \"$got\", expected \"lingqprep v$version\"; is the commit tagged v$version?" >&2
	exit 1
fi

tar -cJf "$out/$name-Linux-$(uname -m).tar.xz" -C "$work" "$name"
echo "wrote $out/$name-Linux-$(uname -m).tar.xz"
