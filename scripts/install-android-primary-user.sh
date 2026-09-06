#!/usr/bin/env bash
# Install the already-built ChatKJB variant only for Android's primary user.
# An existing identity is retained automatically; universal is the default
# when neither identity is installed. Clone/managed profiles are rejected for
# the legacyPhone install so a second app instance cannot be reactivated.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APK_INPUT="${1:-${CHATKJB_APK_OUTPUT_DIR:-$REPO_ROOT/app/app/build/outputs/apk}}"
DEVICE_SERIAL="${2:-}"
LEGACY_PACKAGE="com.neamkim.chatkjb"
UNIVERSAL_PACKAGE="com.termux"

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

reject_secondary_business_profiles() {
  local profiles
  profiles="$(adb_run shell cmd user list -v 2>&1)"
  if grep -Eq 'type=profile\.(CLONE|MANAGED)' <<<"$profiles"; then
    echo "refusing install: Android clone/work profile exists" >&2
    grep -E 'type=profile\.(CLONE|MANAGED)' <<<"$profiles" >&2
    exit 3
  fi
}

installed_for_primary_user() {
  local package_name="$1"
  adb_run shell pm list packages --user 0 "$package_name" |
    grep -qx "package:$package_name"
}

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
  [[ -n "$aapt" ]] || {
    echo "aapt not found; set ANDROID_HOME or add build-tools to PATH" >&2
    exit 2
  }
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

resolve_apk_path() {
  local target_package="$1"
  local candidate package_name version_code exact_path
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

  case "$target_package" in
    "$LEGACY_PACKAGE")
      exact_path="$APK_INPUT/legacyPhone/debug/app-legacyPhone-debug.apk"
      ;;
    "$UNIVERSAL_PACKAGE")
      exact_path="$APK_INPUT/universal/debug/app-universal-debug.apk"
      ;;
  esac
  if [[ -f "$exact_path" ]]; then
    printf '%s\n' "$exact_path"
    return 0
  fi

  while IFS= read -r candidate; do
    package_name="$(apk_package_name "$candidate")" || continue
    [[ "$package_name" == "$target_package" ]] || continue
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
    echo "ambiguous APKs for $target_package at versionCode $best_version under $APK_INPUT" >&2
    printf '  %s\n' "${best_paths[@]}" >&2
    return 8
  fi

  echo "no versioned APK for $target_package found under $APK_INPUT" >&2
  return 5
}

[[ "$(adb_run get-state)" == "device" ]] || {
  echo "target Android device is not ready" >&2
  exit 2
}

AAPT="$(find_aapt)"

if installed_for_primary_user "$LEGACY_PACKAGE" && installed_for_primary_user "$UNIVERSAL_PACKAGE"; then
  echo "refusing install: both ChatKJB identities are installed for user 0" >&2
  exit 6
elif installed_for_primary_user "$LEGACY_PACKAGE"; then
  TARGET_PACKAGE="$LEGACY_PACKAGE"
elif installed_for_primary_user "$UNIVERSAL_PACKAGE"; then
  TARGET_PACKAGE="$UNIVERSAL_PACKAGE"
else
  TARGET_PACKAGE="$UNIVERSAL_PACKAGE"
fi

if [[ "$TARGET_PACKAGE" == "$LEGACY_PACKAGE" ]]; then
  reject_secondary_business_profiles
fi

if [[ -f "$APK_INPUT" ]]; then
  APK_PATH="$APK_INPUT"
else
  APK_PATH="$(resolve_apk_path "$TARGET_PACKAGE")"
fi
APK_PACKAGE="$(apk_package_name "$APK_PATH")"
case "$APK_PACKAGE" in
  "$LEGACY_PACKAGE"|"$UNIVERSAL_PACKAGE") ;;
  *)
    echo "unsupported APK applicationId: ${APK_PACKAGE:-unknown}" >&2
    exit 5
    ;;
esac

if [[ "$APK_PACKAGE" != "$TARGET_PACKAGE" ]]; then
  echo "APK identity $APK_PACKAGE does not match selected installed identity $TARGET_PACKAGE" >&2
  echo "build the matching variant; universal ($UNIVERSAL_PACKAGE) is the new-install default" >&2
  exit 7
fi

adb_run install --user 0 -r "$APK_PATH"
if [[ "$TARGET_PACKAGE" == "$LEGACY_PACKAGE" ]]; then
  reject_secondary_business_profiles
fi

if ! installed_for_primary_user "$TARGET_PACKAGE"; then
  echo "primary-user package verification failed" >&2
  exit 4
fi

echo "installed $TARGET_PACKAGE for primary user 0 only"
