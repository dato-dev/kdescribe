#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <version>" >&2
  echo "example: $0 v1.2.0" >&2
  exit 1
fi

version="$1"
checksums_file="dist/checksums.txt"
manifest_file="krew/kdescribe.yaml"

if [[ ! -f "$checksums_file" ]]; then
  echo "missing $checksums_file; run goreleaser release --snapshot --clean --skip=publish first" >&2
  exit 1
fi

checksum_for() {
  local archive="$1"
  awk -v archive="$archive" '$2 == archive { print $1 }' "$checksums_file"
}

darwin_amd64="$(checksum_for kdescribe_darwin_amd64.tar.gz)"
darwin_arm64="$(checksum_for kdescribe_darwin_arm64.tar.gz)"
linux_amd64="$(checksum_for kdescribe_linux_amd64.tar.gz)"
linux_arm64="$(checksum_for kdescribe_linux_arm64.tar.gz)"
windows_amd64="$(checksum_for kdescribe_windows_amd64.tar.gz)"
windows_arm64="$(checksum_for kdescribe_windows_arm64.tar.gz)"

for value in "$darwin_amd64" "$darwin_arm64" "$linux_amd64" "$linux_arm64" "$windows_amd64" "$windows_arm64"; do
  if [[ -z "$value" ]]; then
    echo "failed to resolve all checksums from $checksums_file" >&2
    exit 1
  fi
done

python3 - "$manifest_file" "$version" \
  "$darwin_amd64" "$darwin_arm64" "$linux_amd64" "$linux_arm64" "$windows_amd64" "$windows_arm64" <<'PY'
from pathlib import Path
import sys

path = Path(sys.argv[1])
version = sys.argv[2]
checksums = sys.argv[3:]
archives = [
    "kdescribe_darwin_amd64.tar.gz",
    "kdescribe_darwin_arm64.tar.gz",
    "kdescribe_linux_amd64.tar.gz",
    "kdescribe_linux_arm64.tar.gz",
    "kdescribe_windows_amd64.tar.gz",
    "kdescribe_windows_arm64.tar.gz",
]

text = path.read_text()
text = __import__("re").sub(r"version: v[0-9]+\.[0-9]+\.[0-9]+", f"version: {version}", text)

for archive, checksum in zip(archives, checksums):
    text = __import__("re").sub(
        rf"(releases/download/)v[0-9]+\.[0-9]+\.[0-9]+(/" + archive + r"\n\s+sha256: )([A-Za-z0-9]+|TODO)",
        rf"\g<1>{version}\g<2>{checksum}",
        text,
    )

path.write_text(text)
PY

echo "updated $manifest_file for $version"
