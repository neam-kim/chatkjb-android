#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd -- "$SCRIPT_DIR/.." && pwd)"
SOURCE_ROOT="${1:-/Volumes/NEAM_SSD/herdr-mobile-relay}"
FRONTEND_ROOT="$SOURCE_ROOT/frontend"
ASSET_ROOT="$PROJECT_ROOT/app/app/src/main/assets/herdr"

if [[ ! -f "$FRONTEND_ROOT/package.json" || ! -f "$SOURCE_ROOT/LICENSE" ]]; then
  echo "usage: $0 [/absolute/path/to/herdr-mobile-relay]" >&2
  exit 2
fi

if [[ "$(git -C "$SOURCE_ROOT" merge-base HEAD origin/main)" != "7400537a429e754fc5af82ee9ac9d28f6e058c1d" ]]; then
  echo "error: embedded frontend must be rebased and reviewed before changing upstream base" >&2
  exit 1
fi

bun run --cwd "$FRONTEND_ROOT" lint
bun run --cwd "$FRONTEND_ROOT" check
bun run --cwd "$FRONTEND_ROOT" test
bun run --cwd "$FRONTEND_ROOT" build
bun run --cwd "$FRONTEND_ROOT" size

mkdir -p "$ASSET_ROOT"
rsync -a --delete \
  --exclude '*.br' \
  --exclude '_headers' \
  --exclude 'SOURCE.txt' \
  "$FRONTEND_ROOT/dist/" "$ASSET_ROOT/"

echo "Embedded ChatKJB frontend synchronized from $SOURCE_ROOT"
