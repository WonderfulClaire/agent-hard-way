#!/usr/bin/env sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
found=0

for module in "$repo_root"/ex*/go.mod; do
  [ -f "$module" ] || continue
  found=1
  chapter=$(dirname "$module")
  printf '==> %s\n' "${chapter#"$repo_root"/}"
  (
    cd "$chapter"
    go test ./...
    go vet ./...
    go build ./...
  )
done

if [ "$found" -eq 0 ]; then
  echo "No exercise modules found." >&2
  exit 1
fi

echo "All exercise modules passed test, vet, and build."
