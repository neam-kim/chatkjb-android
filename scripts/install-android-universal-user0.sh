#!/usr/bin/env bash
# Install the universal ChatKJB package for primary user 0.
# This procedure is intentionally separate from the legacy phone guard: it
# leaves any existing tablet clone profile untouched and never installs into it.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APK_INPUT="${1:-${CHATKJB_APK_OUTPUT_DIR:-$REPO_ROOT/app/app/build/outputs/apk}}"
DEVICE_SERIAL="${2:-}"
PACKAGE_NAME="com.termux"

if [[ "${1:-}" == "--serial" || "${1:-}" == "--" ]]; then
  APK_INPUT="${CHATKJB_APK_OUTPUT_DIR:-$REPO_ROOT/app/app/build/outputs/apk}"
  DEVICE_SERIAL="${2:-}"
fi

if [[ -n "${ADB_BIN:-}" ]]; then
  ADB="$ADB_BIN"
elif [[ -n "${ANDROID_HOME:-}" && -x "${ANDROID_HOME}/platform-tools/adb" ]]; then
  ADB="${ANDROID_HOME}/platform-tools/adb"
elif command -v adb >/dev/null 2>&1; then
  ADB="$(command -v adb)"
else
  echo "adb not found; set ADB_BIN or ANDROID_HOME" >&2
  exit 2
fi

adb_run() {
  if [[ -n "$DEVICE_SERIAL" ]]; then
    "$ADB" -s "$DEVICE_SERIAL" "$@"
  else
    "$ADB" "$@"
  fi
}

if [[ "$(adb_run get-state)" != "device" ]]; then
  echo "target Android device is not ready" >&2
  exit 2
fi

profiles="$(adb_run shell cmd user list -v 2>&1)"
if grep -Eq 'type=profile\.(CLONE|MANAGED)' <<<"$profiles"; then
  echo "clone/managed profile detected; universal install remains user 0 only" >&2
fi

find_aapt() {
  local aapt=""
  if [[ -n "${ANDROID_HOME:-}" ]]; then
    while IFS= read -r candidate; do
      if [[ -x "$candidate/aapt" ]]; then
        aapt="$candidate/aapt"
        break
      fi
    done < <(find "$ANDROID_HOME/build-tools" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | sort -r)
  fi
  if [[ -z "$aapt" ]] && command -v aapt >/dev/null 2>&1; then
    aapt="$(command -v aapt)"
  fi
  [[ -n "$aapt" ]] || { echo "aapt not found; set ANDROID_HOME" >&2; exit 2; }
  printf '%s\n' "$aapt"
}

apk_package_name() {
  local apk_path="$1"
  local aapt_output
  aapt_output="$("$AAPT" dump badging "$apk_path" 2>/dev/null)" || {
    echo "unable to inspect APK: $apk_path" >&2
    return 1
  }
  sed -n "s/^package: name='\([^']*\)'.*/\1/p" <<<"$aapt_output"
}

apk_version_code() {
  local apk_path="$1"
  local aapt_output
  aapt_output="$("$AAPT" dump badging "$apk_path" 2>/dev/null)" || return 1
  sed -n "s/^package:.*versionCode='\([^']*\)'.*/\1/p" <<<"$aapt_output"
}

resolve_universal_apk() {
  local candidate pkg version_code exact_path
  local best_version=0
  local best_count=0
  local -a best_paths=()

  if [[ -f "$APK_INPUT" ]]; then
    printf '%s\n' "$APK_INPUT"
    return 0
  fi
  [[ -d "$APK_INPUT" ]] || {
    echo "APK file or directory not found: $APK_INPUT" >&2
    return 2
  }

  exact_path="$APK_INPUT/universal/debug/app-universal-debug.apk"
  if [[ -f "$exact_path" ]]; then
    printf '%s\n' "$exact_path"
    return 0
  fi

  while IFS= read -r candidate; do
    pkg="$(apk_package_name "$candidate")" || continue
    [[ "$pkg" == "$PACKAGE_NAME" ]] || continue
    version_code="$(apk_version_code "$candidate")" || continue
    [[ "$version_code" =~ ^[0-9]+$ ]] || continue
    if (( best_count == 0 || 10#$version_code > 10#$best_version )); then
      best_version="$version_code"
      best_count=1
      best_paths=("$candidate")
    elif (( 10#$version_code == 10#$best_version )); then
      best_count=$((best_count + 1))
      best_paths+=("$candidate")
    fi
  done < <(find "$APK_INPUT" -type f -name '*.apk' -print | sort)

  if (( best_count == 1 )); then
    printf '%s\n' "${best_paths[0]}"
    return 0
  fi
  if (( best_count > 1 )); then
    echo "ambiguous universal APKs at versionCode $best_version under $APK_INPUT" >&2
    printf '  %s\n' "${best_paths[@]}" >&2
    return 8
  fi

  echo "no versioned universal APK ($PACKAGE_NAME) found under $APK_INPUT" >&2
  return 5
}

AAPT="$(find_aapt)"
APK_PATH="$(resolve_universal_apk)"
APK_PACKAGE="$(apk_package_name "$APK_PATH")"
[[ "$APK_PACKAGE" == "$PACKAGE_NAME" ]] || {
  echo "expected universal APK identity $PACKAGE_NAME, got ${APK_PACKAGE:-unknown}" >&2
  exit 5
}

adb_run install --user 0 -r "$APK_PATH"
if ! adb_run shell pm list packages --user 0 "$PACKAGE_NAME" | grep -qx "package:$PACKAGE_NAME"; then
  echo "universal package verification failed for primary user 0" >&2
  exit 4
fi

echo "installed $PACKAGE_NAME for primary user 0; clone/managed profiles were not modified"
