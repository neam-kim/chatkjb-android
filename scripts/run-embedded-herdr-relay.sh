#!/usr/bin/env bash
set -euo pipefail

ENV_FILE="${HERDR_RELAY_ENV:-${XDG_CONFIG_HOME:-$HOME/.config}/herdr/plugins/config/herdr-mobile-relay.events/relay.env}"
RELAY_BIN="${HERDR_RELAY_BIN:-${XDG_DATA_HOME:-$HOME/.local/share}/herdr-mobile-relay/current/herdr-mobile-relay}"

if [[ ! -r "$ENV_FILE" ]]; then
  echo "ChatKJB relay environment is not readable: $ENV_FILE" >&2
  exit 78
fi
if [[ ! -x "$RELAY_BIN" ]]; then
  echo "Verified Herdr relay binary is not executable: $RELAY_BIN" >&2
  exit 78
fi

set -a
# shellcheck source=/dev/null
. "$ENV_FILE"
set +a

: "${HERDR_RELAY_TOKEN:?Missing HERDR_RELAY_TOKEN in relay environment}"

export PATH="/opt/homebrew/bin:/usr/local/bin:$HOME/.local/bin:/usr/bin:/bin:/usr/sbin:/sbin"
export HERDR_BIN="${HERDR_BIN:-$(command -v herdr)}"
export HERDR_RELAY_HOST="${HERDR_RELAY_HOST:-127.0.0.1}"
export HERDR_RELAY_PORT="${HERDR_RELAY_PORT:-8375}"

TAILNET_RELAY_ORIGIN="${HERDR_TAILNET_RELAY_ORIGIN:-https://neam-macmini.taild81d38.ts.net:8443}"
if command -v tailscale >/dev/null 2>&1 &&
  tailscale serve status 2>/dev/null | grep -Fq "$TAILNET_RELAY_ORIGIN (tailnet only)"; then
  echo "ChatKJB tailnet ingress ready: $TAILNET_RELAY_ORIGIN"
else
  echo "warning: ChatKJB tailnet ingress is not visible at $TAILNET_RELAY_ORIGIN" >&2
fi

exec "$RELAY_BIN" serve
