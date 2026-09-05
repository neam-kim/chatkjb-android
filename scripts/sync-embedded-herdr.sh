#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd -- "$SCRIPT_DIR/.." && pwd)"
SOURCE_ROOT="$PROJECT_ROOT/embedded/herdr-mobile-relay"
FRONTEND_ROOT="$SOURCE_ROOT/frontend"
ASSET_ROOT="$PROJECT_ROOT/app/app/src/main/assets/herdr"

if [[ $# -ne 0 ]]; then
  echo "usage: $0 (builds the repository-owned embedded frontend)" >&2
  exit 2
fi

if [[ ! -f "$FRONTEND_ROOT/package.json" || ! -f "$SOURCE_ROOT/LICENSE" ]]; then
  echo "error: repository-owned Herdr sources are missing" >&2
  exit 1
fi

bun install --cwd "$FRONTEND_ROOT" --frozen-lockfile
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
