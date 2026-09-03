#!/usr/bin/env bash
# Install ChatKJB only for Android's primary user. This intentionally fails
# closed when a clone/work profile exists so a normal `adb install -r` cannot
# silently reactivate a second app instance in that profile.
set -euo pipefail

APK_PATH="${1:-}"
DEVICE_SERIAL="${2:-}"
PACKAGE_NAME="com.neamkim.chatkjb"

if [[ -z "$APK_PATH" || ! -f "$APK_PATH" ]]; then
  echo "usage: $0 /absolute/path/to/app-debug.apk [adb-serial]" >&2
  exit 2
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

reject_secondary_business_profiles() {
  local profiles
  profiles="$(adb_run shell cmd user list -v 2>&1)"
  if grep -Eq 'type=profile\.(CLONE|MANAGED)' <<<"$profiles"; then
    echo "refusing install: Android clone/work profile exists" >&2
    grep -E 'type=profile\.(CLONE|MANAGED)' <<<"$profiles" >&2
    exit 3
  fi
}

[[ "$(adb_run get-state)" == "device" ]] || {
  echo "target Android device is not ready" >&2
  exit 2
}

reject_secondary_business_profiles
adb_run install --user 0 -r "$APK_PATH"
reject_secondary_business_profiles

if ! adb_run shell pm list packages --user 0 "$PACKAGE_NAME" | grep -qx "package:$PACKAGE_NAME"; then
  echo "primary-user package verification failed" >&2
  exit 4
fi

echo "installed $PACKAGE_NAME for primary user 0 only"
