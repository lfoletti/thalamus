#!/bin/sh
# Build the release archives of Thalamus.
#   sh tools/release/build.sh <version> <out-dir>
# For each platform: thalamus_<version>_<os>_<arch>.tar.gz (th-gateway,
# th-hook) and thalamus-plugin_<version>_<os>_<arch>.tar.gz (the Claude Code
# plugin with th-hook in plugin/bin). Once: thalamus-extension_<version>.zip
# (the Chrome extension) and checksums.txt (SHA-256 of each archive).
# The gateway is pure Go, so CGO_ENABLED=0 cross-compiles every platform.
set -eu
version=${1:?version, for example v0.1.0}
out=${2:?output directory}
case $version in
v[0-9]*) ;;
*) echo "build.sh: the version must start with v and a digit, for example v0.1.0: $version" >&2; exit 2 ;;
esac
case $version in
*/*) echo "build.sh: the version must not contain a slash: $version" >&2; exit 2 ;;
esac
root=$(cd "$(dirname "$0")/../.." && pwd)
mkdir -p "$out"
out=$(cd "$out" && pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

for target in darwin/amd64 darwin/arm64 linux/amd64 linux/arm64; do
	os=${target%/*}
	arch=${target#*/}
	dir="$work/thalamus_${version}_${os}_${arch}"
	mkdir -p "$dir"
	for cmd in th-gateway th-hook; do
		(cd "$root/gateway" && CGO_ENABLED=0 GOOS=$os GOARCH=$arch \
			go build -trimpath -ldflags='-s -w' -o "$dir/$cmd" "./cmd/$cmd")
	done
	cp "$root/LICENSE" "$root/README.md" "$dir/"
	tar -C "$work" -czf "$out/thalamus_${version}_${os}_${arch}.tar.gz" "thalamus_${version}_${os}_${arch}"

	plug="$work/plugin_${os}_${arch}/thalamus-plugin"
	mkdir -p "$plug"
	cp -R "$root/plugin/." "$plug/"
	cp "$dir/th-hook" "$plug/bin/th-hook"
	cp "$root/LICENSE" "$plug/"
	tar -C "$work/plugin_${os}_${arch}" -czf "$out/thalamus-plugin_${version}_${os}_${arch}.tar.gz" thalamus-plugin
done

ext="$work/thalamus-extension"
mkdir -p "$ext"
(cd "$root/extension" && cp -R manifest.json background.js lib.js options.html options.js popup.html popup.js icons README.md "$ext/")
cp "$root/LICENSE" "$ext/"
(cd "$ext" && zip -qr "$out/thalamus-extension_${version}.zip" .)

(cd "$out" && rm -f checksums.txt && for f in *.tar.gz *.zip; do
	if command -v sha256sum >/dev/null; then sha256sum "$f"; else shasum -a 256 "$f"; fi
done > checksums.txt)
ls "$out"
